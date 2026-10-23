package ui

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// Single-letter answers are accepted in either case: "A" is [a], "Y" is y.
func TestAnswersIgnoreCase(t *testing.T) {
	var out bytes.Buffer
	u := New(context.Background(), strings.NewReader("A\nE\nY\nN\n\n"), &out)
	for _, want := range []byte{'a', 'e'} {
		if c, err := u.Choose("? ", "aer", 0); err != nil || c != want {
			t.Fatalf("Choose = %q, %v; want %q", c, err, want)
		}
	}
	if ok, err := u.YesNo("?", false); err != nil || !ok {
		t.Fatalf("Y: %v %v", ok, err)
	}
	if ok, err := u.YesNo("?", true); err != nil || ok {
		t.Fatalf("N: %v %v", ok, err)
	}
	if ok, err := u.YesNo("?", true); err != nil || !ok {
		t.Fatalf("Enter must take the default: %v %v", ok, err)
	}
}
