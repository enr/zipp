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
	"github.com/enr/zipext"
	"github.com/mattn/go-colorable"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"

	"github.com/enr/zipp/lib/core"
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

type step struct {
	path        string
	expanded    string
	destination string
	destDir     string
	innerPath   string
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

	params, err := loadParams(c)
	if err != nil {
		return err
	}
	params.DryRun = c.Bool("dry-run")
	params.Force = c.Bool("force")
	params.NoOverwrite = c.Bool("no-overwrite")

	ui.Confidentialf("Adding file=%s to zip=%s in path inner=%s", params.FileToAdd, params.ZipPath, params.InnerPath)

	// Show spinner during the write operation (medium verbosity only)
	stopSpinner := func() {}
	if !c.Bool("quiet") && !c.Bool("verbose") && !params.DryRun {
		stopSpinner = core.StartSpinner(fmt.Sprintf("Updating %s ...", params.ZipPath))
	}

	zipWriter := core.NewZipWriter(ui)
	err = zipWriter.Write(params)
	stopSpinner()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Hint:  %s\n", core.Hint("entry_not_found"))
		return exitErrorf(1, "Error processing %s: %s", params.ZipPath, err.Error())
	}
	return nil
}

func exitErrorf(exitCode int, template string, args ...interface{}) error {
	ui.Errorf(`Something gone wrong:`)
	return cli.NewExitError(fmt.Sprintf(template, args...), exitCode)
}

func loadParams(c *cli.Context) (core.WriterRequest, error) {
	params := core.WriterRequest{}

	fileToAdd := c.String("file")
	innerPath := c.String("inner")
	zipPath := c.String("zip")
	paramsFile := c.String("params")

	ui.Confidentialf("Adding file=%s to zip=%s in path inner=%s using file params=%s", fileToAdd, zipPath, innerPath, paramsFile)

	if fileToAdd == "" && paramsFile == "" {
		upe, _ := ui.QuestionWithDefault("Do you want to use a params file?", "yes")
		upe = strings.ToLower(upe)
		if upe == "yes" || upe == "y" {
			paramsFile, _ = ui.QuestionWithDefault("Which params file?", "zipw.yml")
			yamlFile, err := ioutil.ReadFile(paramsFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Hint:  %s\n", core.Hint("params_file_missing"))
				return params, exitErrorf(2, "Params file not found %s", paramsFile)
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
		cli.StringFlag{Name: "zip, z", Value: "", Usage: "the zip archive to update"},
		cli.StringFlag{Name: "inner, i", Value: "", Usage: "inner path inside the archive (defaults to the file path)"},
		cli.StringFlag{Name: "params, p", Value: "", Usage: "YAML file containing parameters"},
		cli.BoolFlag{Name: "dry-run, N", Usage: "print what would happen without modifying any file"},
		cli.BoolFlag{Name: "force", Usage: "overwrite existing entries without prompting"},
		cli.BoolFlag{Name: "no-overwrite", Usage: "abort with an error if the entry already exists"},
	}

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
