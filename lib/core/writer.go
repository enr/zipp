package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/enr/clui"
	"github.com/enr/go-files/files"
	"github.com/enr/zipext"
)

// WriterRequest represent request for a write operation
type WriterRequest struct {
	FileToAdd string `yaml:"file"`
	InnerPath string `yaml:"inner"`
	ZipPath   string `yaml:"zip"`
}

type step struct {
	path        string
	expanded    string
	destination string
	destDir     string
	innerPath   string
}

// ZipWriter write content to a zip
type ZipWriter struct {
	ui *clui.Clui
}

// NewZipWriter factory function for zip writer component
func NewZipWriter(ui *clui.Clui) *ZipWriter {
	return &ZipWriter{ui: ui}
}

func (w *ZipWriter) Write(params WriterRequest) error {
	if !files.Exists(params.FileToAdd) {
		return fmt.Errorf(`Invalid path for the file to add: "%s". Exit`, params.FileToAdd)
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
	if err != nil {
		return err
	}

	return nil
}

func extractToTmp(zipPath string) (string, error) {
	if !files.IsRegular(zipPath) {
		return "", fmt.Errorf(`Invalid zip file "%s"`, zipPath)
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
