package app

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/z4dm42f5kp-ux/numb-form/internal/config"
	"github.com/z4dm42f5kp-ux/numb-form/internal/formats"
	"github.com/z4dm42f5kp-ux/numb-form/internal/fsutil"
	"github.com/z4dm42f5kp-ux/numb-form/internal/runlog"
)

// isForm reports whether a file name is a risk assessment form: it starts
// with the template's name and has a template extension.
func isForm(name, templateBase string) bool {
	return templateBase != "" && strings.HasPrefix(strings.ToLower(name), strings.ToLower(templateBase)) && formats.IsTemplate(path.Ext(name))
}

// Origins of a form in the working copy.
const (
	byNumbForm = "numb-form" // exactly as numb-form wrote it
	editedForm = "edited"    // not as any recorded run of numb-form wrote it
	otherFile  = ""          // not a form, or not readable
)

// knownForms returns the checksums of the forms numb-form wrote, as the
// manifests of its runs record them: review runs in Downloads, direct
// runs in the log folder. numb-form marks nothing in its documents.
func knownForms(cfg config.Config) map[string]bool {
	var dirs []string
	dl := cfg.Downloads
	if dl == "" {
		dl, _ = fsutil.DownloadsDir()
	}
	if dl != "" {
		dirs = append(dirs, runlog.RunsDir(dl))
	}
	logDir := cfg.LogDir
	if logDir == "" {
		logDir, _ = config.StateDir()
	}
	if logDir != "" {
		dirs = append(dirs, logDir)
	}
	known := map[string]bool{}
	for _, d := range dirs {
		ents, err := os.ReadDir(d)
		if err != nil {
			continue
		}
		for _, e := range ents {
			if !e.IsDir() {
				continue
			}
			m, err := runlog.Load(filepath.Join(d, e.Name()))
			if err != nil {
				continue
			}
			for _, c := range m.Companies {
				for _, f := range c.Files {
					if f.SHA256 != "" {
						known[f.SHA256] = true
					}
				}
			}
		}
	}
	return known
}

// originOf tells whether the file at p is a form exactly as numb-form
// wrote it: its checksum is one that a run of numb-form recorded.
func originOf(p, rel string, rc reviewConfig) string {
	if !isForm(path.Base(rel), rc.templateBase) {
		return otherFile
	}
	h, _, err := fsutil.SHA256File(p)
	switch {
	case err != nil:
		return otherFile
	case rc.known[h]:
		return byNumbForm
	}
	return editedForm
}

func originLabel(o string) string {
	switch o {
	case byNumbForm:
		return "made by numb-form, unchanged"
	case editedForm:
		return "a form edited after filling, made by hand or filled on another computer"
	}
	return ""
}
