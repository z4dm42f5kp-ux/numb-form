package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/formats"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
)

// CheckTemplate reports whether a template can be filled. With a sample
// path it also writes a filled example there (the path must not exist).
// The template itself is only read.
func CheckTemplate(out io.Writer, cfg config.Config, tplPath, sample string) int {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	t, err := formats.LoadTemplate(ctx, tplPath, cfg)
	if err != nil {
		fmt.Fprintln(out, "check-template:", err)
		return ExitError
	}
	warns, err := t.Check()
	if err != nil {
		fmt.Fprintln(out, "check-template: NOT fillable:", err)
		return ExitError
	}
	for _, w := range append(t.Notes, warns...) {
		fmt.Fprintln(out, "note:", w)
	}
	fmt.Fprintln(out, "Template is fillable: section A and all labels were found; section B is never touched.")
	if sample == "" {
		return ExitOK
	}
	if filepath.Ext(sample) != t.OutExt {
		fmt.Fprintf(out, "note: filled forms from this template are %s files\n", t.OutExt)
	}
	vals := form.Values{
		form.FieldName:      "Foo GmbH",
		form.FieldLegalForm: "GmbH",
		form.FieldAddress:   "Foo Street 64, 80333 München, Germany",
		form.FieldReps:      "Jane Doe (Managing Director, Legal Representative), John Roe (Authorized Signatory (Prokurist)) " + CautionText,
		form.FieldDate:      "26/09/2026",
		form.FieldPerson:    "Max Muster",
	}
	filled, _, err := t.Fill(ctx, vals)
	if err != nil {
		fmt.Fprintln(out, "check-template:", err)
		return ExitError
	}
	if err := fsutil.WriteNew(sample, filled); err != nil {
		fmt.Fprintln(out, "check-template:", err)
		return ExitError
	}
	fmt.Fprintln(out, "Sample written to", sample)
	return ExitOK
}

// ShowConfig prints the configuration files and the effective values.
func ShowConfig(out io.Writer, cfg config.Config, notes []string) int {
	fmt.Fprintln(out, "Configuration files, in the order they are read:")
	for _, p := range config.SystemPaths() {
		fmt.Fprintf(out, "  system  %s%s\n", p, exists(p))
	}
	if up, err := config.UserPath(); err == nil {
		fmt.Fprintf(out, "  user    %s%s\n", up, exists(up))
	}
	for _, n := range notes {
		fmt.Fprintln(out, "note:", n)
	}
	fmt.Fprintf(out, "\nmode = %s\nconfirm = %s\nonline = %v\noverview_dir = %s\nheader = %s\ntemplate_name = %s\ntimeout = %s\n",
		cfg.Mode, cfg.Confirm, cfg.Online, cfg.OverviewDir, cfg.Header, cfg.TemplateName, cfg.Timeout)
	fmt.Fprintf(out, "address_sources = %v\nreps_sources = %v\nreps_min_trust = %s\ncaution = %s\nlegacy_formats = %v\npages_write = %v\n",
		cfg.AddressSources, cfg.RepsSources, cfg.RepsMinTrust, cfg.Caution, cfg.LegacyFormats, cfg.PagesWrite)
	for k, o := range cfg.Origin {
		fmt.Fprintf(out, "  (%s set in %s)\n", k, o)
	}
	return ExitOK
}

func exists(p string) string {
	if _, err := os.Stat(p); err == nil {
		return "  (present)"
	}
	return "  (not present)"
}

// InitConfig writes the commented default file to the per-user location.
func InitConfig(out io.Writer) int {
	p, err := config.UserPath()
	if err != nil {
		fmt.Fprintln(out, "config init:", err)
		return ExitError
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		fmt.Fprintln(out, "config init:", err)
		return ExitError
	}
	if err := fsutil.WriteNew(p, []byte(config.Template())); err != nil {
		fmt.Fprintln(out, "config init:", err)
		return ExitError
	}
	fmt.Fprintln(out, "Wrote", p)
	return ExitOK
}
