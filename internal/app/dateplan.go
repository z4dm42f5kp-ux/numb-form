package app

import (
	"fmt"
	"strconv"

	"github.com/z4dm42f5kp-ux/numb-form/internal/reportdate"
	"github.com/z4dm42f5kp-ux/numb-form/internal/ui"
)

// reviewDates lists the dates to write and asks, as numb-form asks for a
// form, whether to write them, edit one or write nothing. ok is false
// when nothing is to be written.
func (r *run) reviewDates(file, column string, plan []planned) ([]planned, bool, error) {
	for len(plan) > 0 {
		r.ui.Printf("\nDates to write into %s (column %q):\n", file, column)
		for i, p := range plan {
			r.ui.Printf("  %d  %-6s %s  %s  (%s)\n", i+1, p.row.DateRef, p.date, ui.Safe(p.row.Name), ui.Safe(p.from))
		}
		ch, err := r.ui.Choose(fmt.Sprintf("Write these %d date(s)? Nothing else in the spreadsheet changes. [a]ccept, [e]dit a date, [r]eject: ", len(plan)), "aer", 0)
		if err != nil {
			return plan, false, err
		}
		switch ch {
		case 'a':
			return plan, true, nil
		case 'r':
			return plan, false, nil
		}
		if plan, err = r.editDate(plan); err != nil {
			return plan, false, err
		}
	}
	return plan, false, nil
}

// editDate changes one date of the plan, or leaves that company out.
func (r *run) editDate(plan []planned) ([]planned, error) {
	a, err := r.ui.Ask(fmt.Sprintf("  Number to change (1-%d, Enter to go back): ", len(plan)))
	if err != nil || a == "" {
		return plan, err
	}
	n, err := strconv.Atoi(a)
	if err != nil || n < 1 || n > len(plan) {
		r.ui.Printf("  Please type a number from 1 to %d.\n", len(plan))
		return plan, nil
	}
	p := &plan[n-1]
	v, err := r.ui.Ask(fmt.Sprintf("  New date for %s (%s, Enter = keep %s, '-' = leave it out): ", ui.Safe(p.row.Name), reportdate.Layout, p.date))
	switch {
	case err != nil || v == "":
		return plan, err
	case v == "-":
		r.logf("date %q: left out by the user", p.row.Name)
		r.skipped = append(r.skipped, p.row.Name+" — left out by you")
		return append(plan[:n-1], plan[n:]...), nil
	}
	d, ok := reportdate.Parse(v)
	if !ok {
		r.ui.Println("  Please type a date such as 16-10-2026.")
		return plan, nil
	}
	r.logf("date %q: %s typed by the user instead of %s", p.row.Name, d, p.date)
	p.date, p.from = d, "typed"
	return plan, nil
}
