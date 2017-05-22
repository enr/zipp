package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/enr/clui"
	"github.com/mattn/go-colorable"
	"github.com/enr/go-files/files"
	"github.com/enr/go-zipext/zipext"
)

const (
	missingParamInputPath = "Oops... I was expecting at least 1 argument: the path to zip."
)

var (
	ui *clui.Clui
	versionTemplate = `%s
Revision: %s
Build date: %s
`
	appVersion = fmt.Sprintf(versionTemplate, version, gitCommit, buildTime)
)

func getUI(level clui.VerbosityLevel) *clui.Clui {
	return &clui.Clui{
		Layout:         &clui.PlainLayout{},
		VerbosityLevel: level,
		Interactive:    true,
		Color:          true,
		Reader:         os.Stdin,
		StdWriter:      colorable.NewColorableStdout(),
		ErrorWriter:    colorable.NewColorableStderr(),
	}
}

func main() {
	runApp(os.Args)
}

func mainAction(c *cli.Context) {
	verbosityLevel := clui.VerbosityLevelMedium
	if c.Bool("verbose") {
		verbosityLevel = clui.VerbosityLevelHigh
	}
	if c.Bool("quiet") {
		verbosityLevel = clui.VerbosityLevelLow
	}
	ui = getUI(verbosityLevel)

	showHelp := func() {
		cli.ShowAppHelp(c)
	}

	fileToAdd := c.String("file")
	innerPath := c.String("inner")
	zipPath := c.String("zip")
	paramsFile := c.String("params")

	ui.Confidentialf("file=%s inner=%s zip=%s params=%s", fileToAdd, innerPath, zipPath, paramsFile)

	if !files.Exists(fileToAdd) {
		ui.Errorf("%s not found. exit", fileToAdd)
		showHelp()
		os.Exit(1)
	}

	tmp, err := extractToTmp(zipPath)
	// clean up
	defer os.RemoveAll(tmp)
	if err != nil {
		fmt.Errorf("error processing %s: %s", zipPath, err.Error())
		os.Exit(1)
	}

	err = addFileToTmp(tmp, fileToAdd, innerPath)
	if err != nil {
		fmt.Errorf("error processing %s: %s", fileToAdd, err.Error())
		os.Exit(1)
	}

	err = zipTmp(tmp, zipPath)
	if err != nil {
		fmt.Errorf("error zipping %s to %s: %s", tmp, zipPath, err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}

func runApp(args []string) {
	app := cli.NewApp()
	app.Name = "zipw"
	app.Usage = "Add files to zip"
	app.Version = appVersion

	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "quiet, q", Usage: "quiet mode"},
		cli.BoolFlag{Name: "verbose, V", Usage: "verbose mode"},
		cli.StringFlag{Name: "file, f", Value: "", Usage: "the file to add"},
		cli.StringFlag{Name: "zip, z", Value: "", Usage: "the zip"},
		cli.StringFlag{Name: "inner, i", Value: "", Usage: "inner path (if missing will be the file path)"},
		cli.StringFlag{Name: "params, p", Value: "", Usage: "the file containing parameters in Yaml format"},
	}

	app.Action = mainAction
	app.Run(args)
}

func extractToTmp(zipPath string) (string, error) {
	dir, err := ioutil.TempDir("", "zipw_")
	if err != nil {
		return "", err
	}
	err = zipext.Extract(zipPath, dir)
	if err != nil {
		return "", err
			fmt.Errorf("error in Extract(%s,%s): %s %s", zipPath, dir, err, err.Error())
	}
	return dir, nil
}

func addFileToTmp(dir string, fileToAdd string, innerPath string) error {
	if len(strings.TrimSpace(innerPath)) == 0 {
		innerPath = fileToAdd
	}
	innerAbsolutePath := path.Join(dir, innerPath)
	innerDir, err := filepath.Abs(filepath.Dir(innerAbsolutePath))
	if err != nil {
		return err
	}
	os.MkdirAll(innerDir, 0755);
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
