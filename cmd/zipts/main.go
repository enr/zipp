// gofmt -tabs=false -tabwidth=2 zip2.go

package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/enr/clui"
	"github.com/enr/go-files/files"
	"github.com/enr/zipext"
	"github.com/mattn/go-colorable"
	"github.com/urfave/cli"

	"github.com/enr/zipp/lib/core"
)

const missingParamInputPath string = "Oops... I was expecting at least 1 argument: the path to zip."

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

type runConfig struct {
	Args      []string
	Noop      bool
	OutputDir string
	Exclude   string
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

	if len(c.Args()) < 1 {
		ui.Error(missingParamInputPath)
		cli.ShowAppHelp(c)
		os.Exit(3)
	}

	showHelp := func() {
		cli.ShowAppHelp(c)
	}
	runConfig := runConfig{
		Args:      c.Args(),
		Noop:      c.Bool("noop"),
		OutputDir: c.String("out"),
		Exclude:   c.String("exclude"),
	}
	os.Exit(run(runConfig, showHelp))
}

func run(c runConfig, showHelp func()) int {
	args := c.Args
	noop := c.Noop
	inputDirPath, err := resolveInputPath(args[0])
	if err != nil {
		ui.Errorf("Error reading input %s : %v", args[0], err)
		return 2
	}
	if !files.Exists(inputDirPath) {
		ui.Errorf("%s not found. exit", inputDirPath)
		if !noop {
			showHelp()
			return 3
		}
	}
	outputDirectory := c.OutputDir
	exclusions := []string{c.Exclude}
	suffix := core.Timestamp()
	targetFilePath, err := resolveOutputPath(inputDirPath, suffix, outputDirectory)
	if noop {
		ui.Lifecyclef("Output file: %s", targetFilePath)
		if err != nil {
			ui.Errorf("Calling zipts in this way you have an error: %v", err)
		}
		return 0
	}
	if err != nil {
		ui.Errorf("Error %v", err)
		return 4
	}
	ui.Confidentialf("Writing to: %s", targetFilePath)
	if noop {
		ui.Warnf("Operating in NOOP mode. Exit without zip")
	} else {
		err := zipext.CreateExcluding(inputDirPath, targetFilePath, exclusions)
		if err != nil {
			ui.Errorf("Error creating %s : %v", targetFilePath, err)
			if files.Exists(targetFilePath) {
				os.Remove(targetFilePath)
			}
			return 5
		}
	}
	if noop {
		ui.Successf("NOOP, skip creating %s", targetFilePath)
	} else {
		ui.Successf("Completed %s", targetFilePath)
	}
	return 0
}

func main() {
	app := cli.NewApp()
	app.Name = "zipts"
	app.Version = appVersion
	app.Usage = "Creates zip archive named after the current timestamp."
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "noop, N", Usage: "noop mode"},
		cli.BoolFlag{Name: "quiet, q", Usage: "quiet mode"},
		cli.BoolFlag{Name: "verbose, V", Usage: "verbose mode"},
		cli.StringFlag{Name: "out, o", Usage: "output directory"},
		cli.StringFlag{Name: "exclude, x", Usage: "exclude path"},
	}
	app.Action = mainAction
	app.Run(os.Args)
}

func resolveInputPath(arg string) (string, error) {
	inputDirPath := strings.TrimSpace(arg)
	if inputDirPath == "" {
		inputDirPath = "."
	}
	inputDirPath, err := filepath.Abs(inputDirPath)
	ui.Lifecyclef("Zipping  %s", inputDirPath)
	if err != nil {
		return "", err
	}
	inputDirPath = filepath.FromSlash(inputDirPath)
	inputDirPath = strings.TrimRight(inputDirPath, string(os.PathSeparator))
	ui.Confidentialf("Resolved input path %s", inputDirPath)
	return inputDirPath, nil
}

func resolveOutputPath(inputPathArg, timestamp, outputDirectoryArg string) (string, error) {
	if outputDirectoryArg == "" {
		outputDirectoryArg = "."
	}
	outputDirectory, err := filepath.Abs(strings.TrimSpace(outputDirectoryArg))
	if err != nil {
		return "", err
	}
	inputPath, err := filepath.Abs(strings.TrimSpace(inputPathArg))
	if err != nil {
		return "", err
	}
	if outputDirectory == inputPath {
		outputDirectory = filepath.Dir(inputPath)
	}
	if !files.IsDir(outputDirectory) {
		return "", fmt.Errorf("Path %s not found. Output directory should exist", outputDirectory)
	}
	baseName := filepath.Base(inputPath)
	targetFilePath := baseName + "-" + timestamp + ".zip"
	targetFilePath = filepath.FromSlash(path.Join(outputDirectory, targetFilePath))
	ui.Confidentialf("Resolved output path: %s", targetFilePath)
	return targetFilePath, nil
}
