package core

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/enr/clui"
	"github.com/enr/go-files/files"
)

func TestRun(t *testing.T) {
	salt := Timestamp()
	readme, _ := filepath.Abs(`../../README.md`)
	testBaseDir, _ := filepath.Abs(`../../test`)
	origFile := filepath.Join(testBaseDir, `data`, `corporate.ear`)
	testFile := filepath.Join(testBaseDir, `work`, fmt.Sprintf(`corporate-%s.ear`, salt))

	ex := files.Exists(testFile)
	if ex {
		err := os.Remove(testFile)
		if err != nil {
			t.Error(err)
		}
	}
	err := copyFile(origFile, testFile, t)
	if err != nil {
		t.Error(err)
	}
	verbosity := func(ui *clui.Clui) {
		ui.VerbosityLevel = clui.VerbosityLevelHigh
	}
	ui, err := clui.NewClui(verbosity)
	if err != nil {
		t.Error(err)
	}
	sut := NewZipWriter(ui)

	innerPathLastToken := fmt.Sprintf(`com/example/readme-%s.md`, salt)
	request := WriterRequest{
		FileToAdd: readme,
		InnerPath: fmt.Sprintf(`webapp.war#WEB-INF/lib/library.jar#%s`, innerPathLastToken),
		ZipPath:   testFile,
	}

	err = sut.Write(request)
	if err != nil {
		t.Error(err)
	}

	tmp, err := extractToTmp(testFile)
	if err != nil {
		t.Error(err)
	}

	wf := path.Join(tmp, `webapp.war`)

	tmp, err = extractToTmp(wf)
	if err != nil {
		t.Error(err)
	}

	lf := path.Join(tmp, `WEB-INF/lib/library.jar`)

	tmp, err = extractToTmp(lf)
	if err != nil {
		t.Error(err)
	}

	actual := path.Join(tmp, innerPathLastToken)
	if !files.Exists(actual) {
		t.Errorf(`No written file %s`, innerPathLastToken)
	}
}

func copyFile(src string, dest string, t *testing.T) error {
	from, err := os.Open(src)
	if err != nil {
		return err
	}
	defer from.Close()

	to, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		return err
	}
	return nil
}
