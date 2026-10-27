package app

import (
	"bytes"
	"strings"
	"testing"
)

// "numb-form man numb-date" shows numb-date's page; "numb-form man
// numb-form", as man.cmd runs it, shows the whole manual.
func TestManTopics(t *testing.T) {
	show := func(args ...string) string {
		t.Helper()
		var b bytes.Buffer
		if code := Man(&b, append(args, "en")); code != ExitOK {
			t.Fatalf("man %v: exit %d\n%s", args, code, b.String())
		}
		return b.String()
	}
	full, date := show(), show("numb-date")
	if !strings.Contains(date, "numb-date writes only .xlsx and .xlsm spreadsheets") || len(date) >= len(full) {
		t.Errorf("man numb-date does not show its page:\n%.400s", date)
	}
	if show("date") != date {
		t.Error("man date differs from man numb-date")
	}
	if show("numb-form") != full {
		t.Error("man numb-form is not the whole manual")
	}
}
