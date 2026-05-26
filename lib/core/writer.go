package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/enr/go-files/files"
	"github.com/enr/zipext"
)

// WriterRequest represents a request for a write operation.
type WriterRequest struct {
	FileToAdd   string `yaml:"file"`
	InnerPath   string `yaml:"inner"`
	ZipPath     string `yaml:"zip"`
	DryRun      bool
	Force       bool
	NoOverwrite bool
}

type step struct {
	path        string
	expanded    string
	destination string
	destDir     string
	innerPath   string
}

// ZipWriter writes content to a zip archive.
type ZipWriter struct {
	ui *UI
}

// NewZipWriter is the factory function for ZipWriter.
func NewZipWriter(ui *UI) *ZipWriter {
	return &ZipWriter{ui: ui}
}

func (w *ZipWriter) Write(params WriterRequest) error {
	if !files.Exists(params.FileToAdd) {
		return fmt.Errorf(`invalid path for the file to add: "%s"`, params.FileToAdd)
	}

	if params.DryRun {
		return w.dryRun(params)
	}

	tmps := []string{}
	defer func() {
		for _, t := range tmps {
			if t != "" {
				w.ui.Confidentialf("Cleanup temporary dir: %s", t)
				os.RemoveAll(t)
			}
		}
	}()

	var err error
	var lastExpanded string
	var lastPath string
	steps := []step{}
	tokens := []string{params.ZipPath}
	tokens = append(tokens, strings.Split(params.InnerPath, `#`)...)
	for i, v := range tokens {
		e := step{}
		e.destDir = lastExpanded
		if i == len(tokens)-1 {
			e.destination = lastPath
			e.path = params.FileToAdd
			e.innerPath = v
			steps = append(steps, e)
			break
		} else if i == 0 {
			e.path = params.ZipPath
			e.destination = params.ZipPath
			e.innerPath = filepath.Base(params.ZipPath)
			e.destDir = filepath.Dir(params.ZipPath)
		} else {
			e.destination = lastPath
			e.path = filepath.Join(lastExpanded, v)
			e.innerPath = v
		}
		lastPath = e.path
		lastExpanded, err = extractToTmp(e.path)
		if err != nil {
			w.ui.Errorf(`%s error extracting %s`, v, e.path)
			break
		}
		tmps = append(tmps, lastExpanded)
		w.ui.Confidentialf("Extracted %s to %s", e.path, lastExpanded)
		e.expanded = lastExpanded
		steps = append(steps, e)
	}
	if err != nil {
		return err
	}

	var s step
	l := len(steps)
	first := true
	for i := l - 1; i > 0; i-- {
		s = steps[i]
		if first {
			if err = w.checkOverwrite(s, params); err != nil {
				return err
			}
			w.ui.Confidentialf("Copy %s in dir %s as %s", s.path, s.destDir, s.innerPath)
			err = addFileToTmp(s.path, s.destDir, s.innerPath)
			if err != nil {
				w.ui.Errorf(`Error copying %s in dir %s as %s: %v`, s.path, s.destDir, s.innerPath, err)
				break
			}
			first = false
		}
		err = zipTmp(s.destDir, s.destination)
		if err != nil {
			w.ui.Errorf(`Error zipping %s to %s: %v`, s.destDir, s.destination, err)
			break
		}
	}
	return err
}

// checkOverwrite verifies whether the target entry already exists and handles
// the overwrite policy (force / no-overwrite / interactive prompt).
func (w *ZipWriter) checkOverwrite(s step, params WriterRequest) error {
	if params.Force {
		return nil
	}
	candidate := path.Join(s.destDir, s.innerPath)
	if !files.Exists(candidate) {
		return nil
	}
	if params.NoOverwrite {
		return fmt.Errorf("entry %q already exists in archive — aborting (use --force to overwrite)", s.innerPath)
	}
	overwrite, err := w.ui.Confirm(
		fmt.Sprintf("Warning: %q already exists in the archive. Overwrite?", s.innerPath),
		false,
	)
	if err != nil {
		return err
	}
	if !overwrite {
		return fmt.Errorf("skipped: entry %q not overwritten", s.innerPath)
	}
	return nil
}

// dryRun prints what would happen without touching any file.
func (w *ZipWriter) dryRun(params WriterRequest) error {
	tokens := []string{params.ZipPath}
	tokens = append(tokens, strings.Split(params.InnerPath, `#`)...)
	finalInnerPath := tokens[len(tokens)-1]

	fmt.Fprintf(os.Stderr, "[dry-run] Would add:     %s\n", params.FileToAdd)
	fmt.Fprintf(os.Stderr, "[dry-run] Destination:   %s :: %s\n", params.ZipPath, finalInnerPath)

	if len(tokens) > 2 {
		nesting := strings.Join(tokens[:len(tokens)-1], " → ")
		fmt.Fprintf(os.Stderr, "[dry-run] Nesting path:  %s\n", nesting)
		fmt.Fprintf(os.Stderr, "[dry-run] Would extract and re-pack each layer.\n")
	}
	fmt.Fprintf(os.Stderr, "[dry-run] No files written.\n")
	return nil
}

func extractToTmp(zipPath string) (string, error) {
	if !files.IsRegular(zipPath) {
		return "", fmt.Errorf(`invalid zip file "%s"`, zipPath)
	}
	dir, err := ioutil.TempDir("", "zipw_")
	if err != nil {
		return "", err
	}
	err = zipext.Extract(zipPath, dir)
	if err != nil {
		return "", err
	}
	return dir, nil
}

func addFileToTmp(fileToAdd string, dir string, innerPath string) error {
	if len(strings.TrimSpace(innerPath)) == 0 {
		innerPath = fileToAdd
	}
	innerAbsolutePath := path.Join(dir, innerPath)
	innerDir, err := filepath.Abs(filepath.Dir(innerAbsolutePath))
	if err != nil {
		return err
	}
	os.MkdirAll(innerDir, 0755)
	err = files.Copy(fileToAdd, innerAbsolutePath)
	if err != nil {
		return err
	}
	return nil
}

func zipTmp(dir string, zipPath string) error {
	err := zipext.CreateFlat(dir, zipPath)
	if err != nil {
		return err
	}
	return nil
}
