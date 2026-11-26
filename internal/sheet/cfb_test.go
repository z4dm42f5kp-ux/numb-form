package sheet

import "testing"

// Sector ids near 2^31 overflow int on 32-bit systems (386, armv7).
// They must be refused, not wrap around to another sector or panic.
func TestCFBLargeSectorIDs(t *testing.T) {
	c := &cfb{data: make([]byte, 2048), ssz: 512, mssz: 64, ministrm: make([]byte, 128), maxChains: 10}
	table := []uint32{cfbEndOfChain, cfbEndOfChain}
	for _, id := range []uint32{0x7FFFFFFE, 0x7FFFFFFF, 0x80000000, 0x80000001, 0xFFFFFFF0} {
		if _, err := c.sector(id); err == nil {
			t.Errorf("sector %#x: no error", id)
		}
		if _, err := c.chain(id, table, true, 64); err == nil {
			t.Errorf("mini chain %#x: no error", id)
		}
		if _, err := c.chain(0, []uint32{id}, true, 128); err == nil {
			t.Errorf("mini chain to %#x: no error", id)
		}
	}
}
