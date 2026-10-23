package ui

import (
	"context"
	"io"
	"strings"
	"testing"
)

// Letters are accepted in either case; Enter takes the default.
func TestChooseIgnoresCase(t *testing.T) {
	u := New(context.Background(), strings.NewReader("A\nY\n\n"), io.Discard)
	for _, want := range []byte{'a', 'y', 'n'} {
		if c, err := u.Choose("? ", "ayn", 'n'); c != want || err != nil {
			t.Errorf("got %q %v, want %q", c, err, want)
		}
	}
}
