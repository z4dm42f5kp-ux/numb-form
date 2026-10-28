// Package runlog writes the human-readable log and the machine-readable
// manifest of a run. Both live in the run directory in Downloads, outside
// the copied tree, so they are never committed into the real directory.
//
// In plain words: every run writes a diary of what it did (the log) and
// a list of the files it made with their checksums (the manifest), so
// that people and the later programs can check them.
package runlog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger appends timestamped lines to a log file.
type Logger struct {
	mu sync.Mutex
	f  *os.File
}

// Open creates or appends to the log file at p.
func Open(p string) (*Logger, error) {
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600) // #nosec G304
	if err != nil {
		return nil, err
	}
	return &Logger{f: f}, nil
}

// Printf logs one line.
func (l *Logger) Printf(format string, a ...any) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.f, "%s  %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, a...))
}

// Close flushes and closes the log.
func (l *Logger) Close() error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := l.f.Sync(); err != nil {
		_ = l.f.Close()
		return err
	}
	return l.f.Close()
}

// File is one file created by a run.
type File struct {
	Rel    string `json:"rel"` // relative to the copied root
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// Company records what happened to one spreadsheet row.
type Company struct {
	Name      string            `json:"name"`
	Cell      string            `json:"cell"`
	Status    string            `json:"status"` // accepted, rejected, exists, skipped, error
	Folder    string            `json:"folder,omitempty"`
	Reason    string            `json:"reason,omitempty"`
	Files     []File            `json:"files,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
	Committed bool              `json:"committed,omitempty"`
}

// Manifest describes a run.
type Manifest struct {
	Version     int       `json:"version"`
	Tool        string    `json:"tool"`
	Mode        string    `json:"mode,omitempty"` // review or direct; empty means review
	Started     time.Time `json:"started"`
	Finished    time.Time `json:"finished,omitempty"`
	SourceRoot  string    `json:"source_root"`
	CopyRoot    string    `json:"copy_root"`
	Spreadsheet string    `json:"spreadsheet"` // relative to root
	SheetSHA256 string    `json:"spreadsheet_sha256"`
	Template    string    `json:"template"`
	Person      string    `json:"person"`
	Companies   []Company `json:"companies"`
	Complete    bool      `json:"complete"`
	Committed   bool      `json:"committed"`
	CommittedAt time.Time `json:"committed_at,omitempty"`
	CopyDeleted bool      `json:"copy_deleted,omitempty"`
	Interrupted string    `json:"interrupted,omitempty"`
}

// ManifestName is the manifest's file name inside a run directory.
const ManifestName = "manifest.json"

// LogName is the run log's file name inside a run directory.
const LogName = "numb-form.log"

// CommitLogName is the commit log's file name inside a run directory.
const CommitLogName = "numb-commit.log"

// Save writes the manifest atomically.
func (m *Manifest) Save(runDir string) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(runDir, ManifestName+".tmp")
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(runDir, ManifestName))
}

// Load reads a manifest.
func Load(runDir string) (*Manifest, error) {
	b, err := os.ReadFile(filepath.Join(runDir, ManifestName)) // #nosec G304
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("manifest: %w", err)
	}
	if m.Version != 1 {
		return nil, fmt.Errorf("manifest version %d not supported", m.Version)
	}
	return &m, nil
}

// RunsDir returns the directory holding all runs inside Downloads.
func RunsDir(downloads string) string { return filepath.Join(downloads, "numb-form-runs") }
