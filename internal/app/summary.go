// The summary at the end of a run.

package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
)

func (r *run) summary() {
	counts := map[string]int{}
	for _, c := range r.man.Companies {
		counts[c.Status]++
	}
	r.ui.Printf("\nDone: %d accepted, %d already had a folder, %d rejected, %d skipped, %d errors.\n",
		counts["accepted"], counts["exists"], counts["rejected"], counts["skipped"], counts["error"])
	if n := byYes(r.man.Companies); n > 0 {
		r.ui.Printf("%d of those folders were taken with -y because their names are similar; check them in the log.\n", n)
	}
	if r.cfg.Mode == "review" {
		r.ui.Printf("Review copy: %s\n", r.workRoot)
	}
	r.ui.Printf("Log:         %s\n", filepath.Join(r.runDir, runlog.LogName))
	if r.cfg.Mode == "review" && counts["accepted"] > 0 {
		r.ui.Println("\nNothing has been written to the real folder. After reviewing, run numb-commit to copy the accepted folders back.")
	}
	r.logf("summary %v", counts)
}

// where describes a spreadsheet row for people: "Customer ID D1001, row
// 315", or "row 315" without an id.
func where(idHeader, id string, row int) string {
	if id == "" {
		return fmt.Sprintf("row %d", row)
	}
	return fmt.Sprintf("%s %s, row %d", idHeader, id, row)
}

// byY starts the reason of a similar folder taken with -y.
const byY = "similar folder, accepted with -y: "

func byYes(cs []runlog.Company) int {
	n := 0
	for _, c := range cs {
		if c.Status == "exists" && strings.HasPrefix(c.Reason, byY) {
			n++
		}
	}
	return n
}

func rel(root, p string) string {
	r, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return r
}
