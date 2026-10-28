package app

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/sheet"
)

// reviewSetup reads what the scan and the review need from the working
// copy: the spreadsheet's companies and ids, and the checksums of the
// forms numb-form wrote.
func reviewSetup(ctx context.Context, from, to string, cfg config.Config) (rc reviewConfig, err error) {
	overview, err := findOverview(from, cfg.OverviewDir)
	if err != nil {
		return rc, err
	}
	sp, _, err := findSpreadsheet(overview)
	if err != nil {
		return rc, err
	}
	tpl, err := findTemplate(from, cfg.TemplateName)
	if err != nil {
		return rc, err
	}
	rc = reviewConfig{header: cfg.Header, idHeader: cfg.IDHeader, overviewDir: cfg.OverviewDir, sheetRel: rel(from, sp),
		template: filepath.ToSlash(rel(from, tpl)), templateBase: strings.TrimSuffix(filepath.Base(tpl), filepath.Ext(tpl)),
		known: knownForms(cfg), names: map[string]string{}, ids: map[string]string{}}
	if b, err := readSheetFile(sp); err == nil {
		if rows, _, err := sheet.Read(b, filepath.Ext(sp), sheet.Options{Header: cfg.Header, IDHeader: cfg.IDHeader, Legacy: cfg.LegacyFormats}); err == nil {
			for _, r := range rows {
				n, _ := fsutil.SafeName(r.Name, false)
				rc.names[strings.ToLower(n)] = r.Name
				rc.ids[strings.ToLower(n)] = r.CustomerID
			}
		}
	}
	return rc, nil
}
