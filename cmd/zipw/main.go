package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/enr/clui"
	"github.com/enr/go-files/files"
	"github.com/enr/go-zipext/zipext"
	"github.com/mattn/go-colorable"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"

	"github.com/enr/zipp/lib/core"
)

const (
	missingParamInputPath = "Oops... I was expecting at least 1 argument: the path to zip."
)

var (
	ui              *clui.Clui
	versionTemplate = `%s
Revision: %s
Build date: %s
`
	appVersion = fmt.Sprintf(versionTemplate, core.Version, core.GitCommit, core.BuildTime)
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

type commandParams struct {
	FileToAdd string `yaml:"file"`
	InnerPath string `yaml:"inner"`
	ZipPath   string `yaml:"zip"`
}

func mainAction(c *cli.Context) error {
	verbosityLevel := clui.VerbosityLevelMedium
	if c.Bool("verbose") {
		verbosityLevel = clui.VerbosityLevelHigh
	}
	if c.Bool("quiet") {
		verbosityLevel = clui.VerbosityLevelLow
	}
	ui = getUI(verbosityLevel)

	// showHelp := func() {
	// 	cli.ShowAppHelp(c)
	// }

	params, err := loadParams(c)
	if err != nil {
		return err
	}
	ui.Confidentialf("file=%s inner=%s zip=%s", params.FileToAdd, params.InnerPath, params.ZipPath)

	if !files.Exists(params.FileToAdd) {
		return exitErrorf(1, `Invalid path "%s". Exit`, params.FileToAdd)
	}

	tmp, err := extractToTmp(params.ZipPath)
	// clean up
	defer os.RemoveAll(tmp)
	if err != nil {
		return exitErrorf(1, "error processing %s: %s", params.ZipPath, err.Error())
	}

	err = addFileToTmp(tmp, params.FileToAdd, params.InnerPath)
	if err != nil {
		return exitErrorf(1, "error processing %s: %s", params.FileToAdd, err.Error())
	}

	err = zipTmp(tmp, params.ZipPath)
	if err != nil {
		return exitErrorf(1, "error zipping %s to %s: %s", tmp, params.ZipPath, err.Error())
	}

	return nil
}

func exitErrorf(exitCode int, template string, args ...interface{}) error {
	ui.Errorf(`Something gone wrong:`)
	return cli.NewExitError(fmt.Sprintf(template, args...), exitCode)
}

func loadParams(c *cli.Context) (commandParams, error) {
	params := commandParams{}

	fileToAdd := c.String("file")
	innerPath := c.String("inner")
	zipPath := c.String("zip")
	paramsFile := c.String("params")

	ui.Confidentialf("file=%s inner=%s zip=%s params=%s", fileToAdd, innerPath, zipPath, paramsFile)

	if fileToAdd == "" && paramsFile == "" {
		upe, _ := ui.QuestionWithDefault("Do you want to use a params file?", "yes")
		upe = strings.ToLower(upe)
		if upe == "yes" || upe == "y" {
			paramsFile, _ = ui.QuestionWithDefault("Which params file?", "zipw.yml")
			yamlFile, err := ioutil.ReadFile(paramsFile)
			if err != nil {
				return params, exitErrorf(1, "yamlFile.Get err   #%v ", err)
			}
			err = yaml.Unmarshal(yamlFile, &params)
			if err != nil {
				return params, exitErrorf(1, "Unmarshal: %v", err)
			}
		} else {
			params.FileToAdd = fileToAdd
			params.InnerPath = innerPath
			params.ZipPath = zipPath
		}
	} else {
		params.FileToAdd = fileToAdd
		params.InnerPath = innerPath
		params.ZipPath = zipPath
	}
	return params, nil
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

	//app.Before = altsrc.InitInputSourceWithContext(flags, altsrc.NewYamlSourceFromFlagFunc("params, p"))
	app.Action = mainAction
	app.Run(args)
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

func addFileToTmp(dir string, fileToAdd string, innerPath string) error {
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
