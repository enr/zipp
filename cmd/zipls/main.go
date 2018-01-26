package main

import (
	"archive/zip"
	"fmt"
	"os"
	"strings"

	"github.com/enr/go-files/files"
	"github.com/enr/go-zipext/zipext"
	"github.com/mattn/go-colorable"
	"github.com/mitchellh/colorstring"
	"github.com/urfave/cli"

	"github.com/enr/zipp/lib/core"
)

const (
	dirTemplate           = "%s\n"
	lineTemplate          = "%s\n"
	lineSizeTemplate      = "%s (%s)\n"
	missingParamInputPath = "Oops... I was expecting at least 1 argument: the path to zip."
)

var (
	versionTemplate = `%s
Revision: %s
Build date: %s
`
	stdout     = colorable.NewColorableStdout()
	stderr     = colorable.NewColorableStderr()
	grepColor  = ""
	appVersion = fmt.Sprintf(versionTemplate, core.Version, core.GitCommit, core.BuildTime)
)

func main() {
	runApp(os.Args)
}

func mainAction(c *cli.Context) error {
	if len(c.Args()) < 1 {
		cli.ShowAppHelp(c)
		return cli.NewExitError(missingParamInputPath, 3)
	}
	grepColor = c.String("grep")

	fileName := c.Args()[0]

	if !files.Exists(fileName) {
		cli.ShowAppHelp(c)
		return cli.NewExitError(fmt.Sprintf("Invalid file: %s", fileName), 2)
	}

	err := zipext.Walk(fileName, func(f *zip.File, err error) error {
		if err != nil {
			fmt.Fprintf(stderr, "Error reading file %s\n", err)
			return err
		}
		if entryRejected(c, f) {
			return nil
		}
		printEntry(c, f)
		return nil
	})

	if err != nil {
		return cli.NewExitError(fmt.Sprintf("error processing %s: %s", fileName, err.Error()), 1)
	}
	return nil
}

func runApp(args []string) {
	app := cli.NewApp()
	app.Name = "zipls"
	app.Usage = "Read zip files"
	app.Version = appVersion

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "grep, g", Value: "", Usage: "grep"},
		cli.StringFlag{Name: "exclude, e", Value: "", Usage: "similar to grep -v"},
		cli.BoolFlag{Name: "dir, d", Usage: "show dirs"},
		cli.BoolFlag{Name: "size, s", Usage: "show size"},
		//cli.BoolFlag{Name: "ignore-case, i", Usage: "ignore case in grep"},
	}

	app.Action = mainAction
	app.Run(args)
}

func printEntry(c *cli.Context, f *zip.File) {
	if entryIsDirectory(f) {
		printDir(f)
		return
	}
	if c.Bool("size") {
		printLineWithSize(f)
		return
	}
	printLine(f)
}

func printLineWithSize(f *zip.File) {
	fmt.Fprintf(stdout, lineSizeTemplate, color(f.Name), humanSize(f.UncompressedSize64))
}
func printLine(f *zip.File) {
	fmt.Fprintf(stdout, lineTemplate, color(f.Name))
}

func printDir(f *zip.File) {
	fmt.Fprintf(stdout, dirTemplate, color(f.Name))
}

func color(str string) string {
	if grepColor == "" {
		return str
	}
	return strings.Replace(str, grepColor, colorstring.Color("[green]"+grepColor), -1) //
}

func entryIsDirectory(f *zip.File) bool {
	return strings.HasSuffix(f.Name, "/") && f.UncompressedSize64 == 0
}

func entryIsRegularFile(f *zip.File) bool {
	return !entryIsDirectory(f)
}

func entryRejected(c *cli.Context, f *zip.File) bool {
	grepWord := c.String("grep")
	if grepWord != "" && !strings.Contains(f.Name, grepWord) {
		return true
	}
	grepvWord := c.String("exclude")
	if grepvWord != "" && strings.Contains(f.Name, grepvWord) {
		return true
	}
	if c.Bool("dir") && entryIsRegularFile(f) {
		return true
	}
	return false
}

// HumanSize returns a human-readable approximation of a size
// using SI standard (eg. "44kB", "17MB")
func humanSize(size uint64) string {
	i := 0
	var sizef float64
	sizef = float64(size)
	units := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	for sizef >= 1000.0 {
		sizef = sizef / 1000.0
		i++
	}
	return fmt.Sprintf("%.4g %s", sizef, units[i])
}
