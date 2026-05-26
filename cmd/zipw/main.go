package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/urfave/cli"
	"gopkg.in/yaml.v3"

	"github.com/enr/zipp/lib/core"
)

var (
	ui              *core.UI
	versionTemplate = `%s
Revision: %s
Build date: %s
`
	appVersion = fmt.Sprintf(versionTemplate, core.Version, core.GitCommit, core.BuildTime)
)

func main() {
	runApp(os.Args)
}

func mainAction(c *cli.Context) error {
	verbosity := core.VerbosityMedium
	if c.Bool("verbose") {
		verbosity = core.VerbosityVerbose
	}
	if c.Bool("quiet") {
		verbosity = core.VerbosityQuiet
	}
	ui = core.NewUI(verbosity)

	params, err := loadParams(c)
	if err != nil {
		return err
	}
	params.DryRun = c.Bool("dry-run")
	params.Force = c.Bool("force")
	params.NoOverwrite = c.Bool("no-overwrite")

	ui.Confidentialf("Adding file=%s to zip=%s in path inner=%s", params.FileToAdd, params.ZipPath, params.InnerPath)

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
		useParamsFile, _ := ui.Confirm("Do you want to use a params file?", true)
		if useParamsFile {
			paramsFile, _ = ui.Input("Which params file?", "zipw.yml", "zipw.yml")
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
