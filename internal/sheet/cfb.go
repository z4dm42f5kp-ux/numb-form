package sheet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"unicode/utf16"
)

// Compound File Binary (MS-CFB) reader, just enough to extract one stream
// from a legacy .xls file. Every offset and chain is bounds-checked; the
// file is untrusted input.

var cfbMagic = []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}

const (
	cfbFree       = 0xFFFFFFFF
	cfbEndOfChain = 0xFFFFFFFE
	cfbMaxRegSect = 0xFFFFFFFA
)

type cfb struct {
	data      []byte
	ssz       int // sector size
	mssz      int // mini sector size
	cutoff    uint32
	fat       []uint32
	minifat   []uint32
	ministrm  []byte
	dirs      []cfbDir
	maxChains int
}

type cfbDir struct {
	name  string
	typ   byte
	start uint32
	size  uint64
}

func u32(b []byte, off int) uint32 { return binary.LittleEndian.Uint32(b[off:]) }
func u16(b []byte, off int) uint16 { return binary.LittleEndian.Uint16(b[off:]) }

func (c *cfb) sector(id uint32) ([]byte, error) {
	if id > cfbMaxRegSect {
		return nil, fmt.Errorf("cfb: bad sector id %#x", id)
	}
	off := (int(id) + 1) * c.ssz
	if off < 0 || off+c.ssz > len(c.data) {
		return nil, fmt.Errorf("cfb: sector %d outside the file", id)
	}
	return c.data[off : off+c.ssz], nil
}

// chain follows a FAT (or mini FAT) chain and concatenates the sectors.
func (c *cfb) chain(start uint32, table []uint32, mini bool, limit uint64) ([]byte, error) {
	var out bytes.Buffer
	seen := map[uint32]bool{}
	for id := start; id != cfbEndOfChain; {
		if seen[id] || len(seen) > c.maxChains {
			return nil, errors.New("cfb: sector chain loops")
		}
		seen[id] = true
		if mini {
			off := int(id) * c.mssz
			if int(id) < 0 || off+c.mssz > len(c.ministrm) {
				return nil, errors.New("cfb: mini sector outside the mini stream")
			}
			out.Write(c.ministrm[off : off+c.mssz])
		} else {
			s, err := c.sector(id)
			if err != nil {
				return nil, err
			}
			out.Write(s)
		}
		if uint64(out.Len()) >= limit {
			break
		}
		if int(id) >= len(table) {
			return nil, errors.New("cfb: chain leaves the allocation table")
		}
		id = table[id]
	}
	b := out.Bytes()
	if uint64(len(b)) < limit {
		return nil, errors.New("cfb: stream shorter than its declared size")
	}
	return b[:limit], nil
}

func openCFB(data []byte) (*cfb, error) {
	if len(data) < 512 || !bytes.Equal(data[:8], cfbMagic) {
		return nil, errors.New("not a legacy Office file (compound file signature missing)")
	}
	if u16(data, 0x1C) != 0xFFFE {
		return nil, errors.New("cfb: bad byte order mark")
	}
	shift, mshift := u16(data, 0x1E), u16(data, 0x20)
	if (shift != 9 && shift != 12) || mshift != 6 {
		return nil, errors.New("cfb: unsupported sector size")
	}
	c := &cfb{data: data, ssz: 1 << shift, mssz: 1 << mshift, cutoff: u32(data, 0x38)}
	c.maxChains = len(data)/c.mssz + 1
	nfat := int(u32(data, 0x2C))
	if nfat < 0 || nfat > len(data)/c.ssz+1 {
		return nil, errors.New("cfb: bad FAT size")
	}
	// FAT sector ids: 109 in the header, the rest in the DIFAT chain.
	var fatIDs []uint32
	for i := 0; i < 109 && len(fatIDs) < nfat; i++ {
		fatIDs = append(fatIDs, u32(data, 0x4C+4*i))
	}
	next := u32(data, 0x44)
	for hops := 0; len(fatIDs) < nfat && next != cfbEndOfChain && next != cfbFree; hops++ {
		if hops > nfat {
			return nil, errors.New("cfb: DIFAT chain loops")
		}
		s, err := c.sector(next)
		if err != nil {
			return nil, err
		}
		per := c.ssz/4 - 1
		for i := 0; i < per && len(fatIDs) < nfat; i++ {
			fatIDs = append(fatIDs, u32(s, 4*i))
		}
		next = u32(s, c.ssz-4)
	}
	for _, id := range fatIDs {
		s, err := c.sector(id)
		if err != nil {
			return nil, err
		}
		for i := 0; i < c.ssz/4; i++ {
			c.fat = append(c.fat, u32(s, 4*i))
		}
	}
	if err := c.readDirs(u32(data, 0x30)); err != nil {
		return nil, err
	}
	root := c.dirs[0]
	if root.typ != 5 {
		return nil, errors.New("cfb: first directory entry is not the root")
	}
	if nmf := u32(data, 0x40); nmf > 0 {
		mf, err := c.chain(u32(data, 0x3C), c.fat, false, uint64(nmf)*uint64(c.ssz))
		if err != nil {
			return nil, err
		}
		for i := 0; i+4 <= len(mf); i += 4 {
			c.minifat = append(c.minifat, u32(mf, i))
		}
	}
	if root.size > 0 {
		if root.size > uint64(len(data)) {
			return nil, errors.New("cfb: mini stream larger than the file")
		}
		ms, err := c.chain(root.start, c.fat, false, root.size)
		if err != nil {
			return nil, err
		}
		c.ministrm = ms
	}
	return c, nil
}

func (c *cfb) readDirs(start uint32) error {
	var raw bytes.Buffer
	seen := map[uint32]bool{}
	for id := start; id != cfbEndOfChain; {
		if seen[id] || len(seen) > c.maxChains {
			return errors.New("cfb: directory chain loops")
		}
		seen[id] = true
		s, err := c.sector(id)
		if err != nil {
			return err
		}
		raw.Write(s)
		if int(id) >= len(c.fat) {
			return errors.New("cfb: directory chain leaves the FAT")
		}
		id = c.fat[id]
	}
	b := raw.Bytes()
	for off := 0; off+128 <= len(b); off += 128 {
		e := b[off : off+128]
		n := int(u16(e, 0x40))
		if n > 64 || n%2 != 0 {
			n = 0
		}
		u := make([]uint16, 0, n/2)
		for i := 0; i+1 < n; i += 2 {
			if v := u16(e, i); v != 0 {
				u = append(u, v)
			}
		}
		c.dirs = append(c.dirs, cfbDir{
			name:  string(utf16.Decode(u)),
			typ:   e[0x42],
			start: u32(e, 0x74),
			size:  uint64(u32(e, 0x78)), // high half is ignored, as v3 files require
		})
	}
	if len(c.dirs) == 0 {
		return errors.New("cfb: no directory entries")
	}
	return nil
}

// stream returns the named stream's contents.
func (c *cfb) stream(name string) ([]byte, error) {
	for _, d := range c.dirs {
		if d.typ != 2 || d.name != name {
			continue
		}
		if d.size > uint64(len(c.data)) {
			return nil, errors.New("cfb: stream larger than the file")
		}
		if d.size < uint64(c.cutoff) {
			return c.chain(d.start, c.minifat, true, d.size)
		}
		return c.chain(d.start, c.fat, false, d.size)
	}
	return nil, fmt.Errorf("stream %q not found", name)
}
