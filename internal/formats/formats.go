// Package formats chooses how a template or spreadsheet is read, filled
// and written, by file extension and configuration.
package formats

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/convert"
	"github.com/z4dm42f5kp-ux/numb-form/internal/docx"
	"github.com/z4dm42f5kp-ux/numb-form/internal/form"
	"github.com/z4dm42f5kp-ux/numb-form/internal/odt"
	"github.com/z4dm42f5kp-ux/numb-form/internal/rtf"
	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
)

// TemplateExts lists template extensions, preferred first.
var TemplateExts = []string{".docx", ".odt", ".fodt", ".rtf", ".pages", ".doc"}

// SheetExts lists spreadsheet extensions.
var SheetExts = append(append([]string(nil), sheet.Extensions...), ".numbers")

func has(list []string, ext string) bool {
	ext = strings.ToLower(ext)
	for _, e := range list {
		if e == ext {
			return true
		}
	}
	return false
}

// IsTemplate reports whether ext is a template extension.
func IsTemplate(ext string) bool { return has(TemplateExts, ext) }

// IsSheet reports whether ext is a spreadsheet extension.
func IsSheet(ext string) bool { return has(SheetExts, ext) }

// Template is a loaded template, ready to fill.
type Template struct {
	Name   string
	OutExt string // extension of the filled files
	Legacy bool   // a legacy format: the run must stay offline
	Notes  []string

	filler  form.Filler
	data    []byte
	finish  func(context.Context, []byte) ([]byte, error)
	toCheck func(context.Context, []byte) ([]byte, error)
}

func native(ext string) form.Filler {
	switch strings.ToLower(ext) {
	case ".docx":
		return docx.Format{}
	case ".odt":
		return odt.Format{}
	case ".fodt":
		return odt.Format{Flat: true}
	case ".rtf":
		return rtf.Format{}
	}
	return nil
}

// ReadFilled reads section A from a filled form in a format numb-form
// reads itself (.docx, .odt, .fodt, .rtf).
func ReadFilled(b []byte, ext string) (form.Values, error) {
	f := native(ext)
	if f == nil {
		return nil, fmt.Errorf("%s forms cannot be read directly", ext)
	}
	return f.ReadBack(b)
}

// LoadTemplate reads and, if needed, converts the template at p.
func LoadTemplate(ctx context.Context, p string, cfg config.Config) (*Template, error) {
	data, err := os.ReadFile(p) // #nosec G304 -- the user's own template.
	if err != nil {
		return nil, err
	}
	if len(data) > 50<<20 {
		return nil, errors.New("template is larger than 50 MB")
	}
	ext := strings.ToLower(filepath.Ext(p))
	t := &Template{Name: filepath.Base(p), OutExt: ext}
	if f := native(ext); f != nil {
		t.filler, t.data = f, data
		return t, nil
	}
	switch ext {
	case ".doc":
		return loadDoc(ctx, t, data, cfg)
	case ".pages":
		return loadPages(ctx, t, data, cfg)
	}
	return nil, fmt.Errorf("unsupported template format %q", ext)
}

func loadDoc(ctx context.Context, t *Template, data []byte, cfg config.Config) (*Template, error) {
	if !cfg.LegacyFormats {
		return nil, errors.New(".doc is a legacy format; an administrator must set legacy_formats = yes in the system configuration")
	}
	so, err := convert.FindSoffice(cfg.Soffice)
	if err != nil {
		return nil, err
	}
	d, err := convert.Soffice(ctx, so, data, ".doc", "docx")
	if err != nil {
		return nil, fmt.Errorf("convert .doc template: %w", err)
	}
	t.filler, t.data, t.Legacy, t.OutExt = docx.Format{}, d, true, ".docx"
	t.Notes = append(t.Notes, "legacy .doc template converted with "+filepath.Base(so)+"; online lookups are off for this run")
	if cfg.LegacyOutput == "same" {
		t.OutExt = ".doc"
		t.finish = func(ctx context.Context, b []byte) ([]byte, error) {
			return convert.Soffice(ctx, so, b, ".docx", "doc")
		}
		t.toCheck = func(ctx context.Context, b []byte) ([]byte, error) {
			return convert.Soffice(ctx, so, b, ".doc", "docx")
		}
	}
	return t, nil
}

func loadPages(ctx context.Context, t *Template, data []byte, cfg config.Config) (*Template, error) {
	useApp := cfg.PagesConverter == "pages-app" || (cfg.PagesConverter == "auto" && convert.HasApp("Pages"))
	var d []byte
	var err error
	if useApp {
		d, err = convert.Apple(ctx, "pages-to-docx", data, ".pages", ".docx")
	} else {
		var so string
		if so, err = convert.FindSoffice(cfg.Soffice); err == nil {
			d, err = convert.Soffice(ctx, so, data, ".pages", "docx")
		}
	}
	if err != nil {
		return nil, fmt.Errorf("convert .pages template: %w", err)
	}
	t.filler, t.data, t.OutExt = docx.Format{}, d, ".docx"
	t.Notes = append(t.Notes, "the .pages template was converted to Word format; check the layout of the first filled form")
	if cfg.PagesWrite {
		if runtime.GOOS != "darwin" || !convert.HasApp("Pages") {
			return nil, errors.New("pages_write needs macOS with Apple Pages")
		}
		t.OutExt = ".pages"
		t.Notes = append(t.Notes, "EXPERIMENTAL: filled forms are saved as .pages through Apple Pages")
		t.finish = func(ctx context.Context, b []byte) ([]byte, error) {
			return convert.Apple(ctx, "docx-to-pages", b, ".docx", ".pages")
		}
		t.toCheck = func(ctx context.Context, b []byte) ([]byte, error) {
			return convert.Apple(ctx, "pages-to-docx", b, ".pages", ".docx")
		}
	}
	return t, nil
}

// Check validates that the template can be filled.
func (t *Template) Check() ([]string, error) { return form.Check(t.filler, t.data) }

// Fill fills the template and converts the result to OutExt.
func (t *Template) Fill(ctx context.Context, v form.Values) ([]byte, []string, error) {
	out, warns, err := t.filler.Fill(t.data, v)
	if err != nil || t.finish == nil {
		return out, warns, err
	}
	final, err := t.finish(ctx, out)
	return final, warns, err
}

// ReadBack reads section A from a filled file, converting it back first
// when the output format is not edited directly.
func (t *Template) ReadBack(ctx context.Context, out []byte) (form.Values, error) {
	if t.toCheck != nil {
		b, err := t.toCheck(ctx, out)
		if err != nil {
			return nil, fmt.Errorf("read back: %w", err)
		}
		out = b
	}
	return t.filler.ReadBack(out)
}

// LoadSheet reads the companies from the spreadsheet at p. legacy is
// true for formats that require an offline run.
func LoadSheet(ctx context.Context, p string, cfg config.Config) (rows []sheet.Row, warns []string, legacy bool, err error) {
	opt := sheet.Options{Header: cfg.Header, IDHeader: cfg.IDHeader, Legacy: cfg.LegacyFormats}
	ext := strings.ToLower(filepath.Ext(p))
	if ext != ".numbers" {
		rows, warns, err = sheet.ReadFile(p, opt)
		return rows, warns, sheet.IsLegacy(ext), err
	}
	data, err := os.ReadFile(p) // #nosec G304 -- the user's own spreadsheet, read only.
	if err != nil {
		return nil, nil, false, err
	}
	var x []byte
	if cfg.NumbersConverter == "numbers-app" || (cfg.NumbersConverter == "auto" && convert.HasApp("Numbers")) {
		x, err = convert.Apple(ctx, "numbers-to-xlsx", data, ".numbers", ".xlsx")
	} else {
		var so string
		if so, err = convert.FindSoffice(cfg.Soffice); err == nil {
			x, err = convert.Soffice(ctx, so, data, ".numbers", "xlsx")
		}
	}
	if err != nil {
		return nil, nil, false, fmt.Errorf("convert .numbers spreadsheet: %w", err)
	}
	rows, warns, err = sheet.Read(x, ".xlsx", opt)
	return rows, append(warns, "the .numbers spreadsheet was converted in memory; the original is untouched"), false, err
}

// AsFilled reports whether doc is this template filled by numb-form and
// unchanged since. Only templates filled directly (.docx, .odt, .fodt,
// .rtf) can tell; for the others it returns an error.
func (t *Template) AsFilled(doc []byte) (bool, error) {
	if t.finish != nil || t.toCheck != nil {
		return false, errors.New("forms of converted templates cannot be compared")
	}
	return form.AsFilled(t.filler, t.data, doc)
}
