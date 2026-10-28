package textutil

import (
	"strings"
	"sync"
	"sync/atomic"
)

// fullUnicode is the "unicode = full" setting. It is off by default; the
// table below is then never read, so the default costs nothing.
var fullUnicode atomic.Bool

// SetFullUnicode turns the full Unicode comparison of names on or off.
//
// In plain words: with it on, letters from every alphabet that carry an
// accent or are written in a special way (ș, ğ, ǎ, ﬁ, full-width Ａ) count
// as their plain letters when names are compared, not only the common
// Western European ones.
func SetFullUnicode(on bool) { fullUnicode.Store(on) }

// fullFold reads unifoldTable into a map the first time it is needed.
var fullFold = sync.OnceValue(func() map[rune]string {
	m := make(map[rune]string, 1600)
	for _, item := range strings.Split(unifoldTable, ",") {
		k, v, ok := strings.Cut(item, " ")
		if !ok {
			continue
		}
		r := []rune(k)
		if len(r) == 1 {
			m[r[0]] = v
		}
	}
	return m
})
