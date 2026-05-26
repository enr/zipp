package main

import (
	"archive/zip"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/enr/go-files/files"
	"github.com/enr/zipext"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli"

	"github.com/enr/zipp/lib/core"
)

const missingParamInputPath = "Oops... I was expecting at least 1 argument: the path to zip."

var (
	versionTemplate = `%s
Revision: %s
Build date: %s
`
	stdout     = os.Stdout
	stderr     = os.Stderr
	grepColor  = ""
	appVersion = fmt.Sprintf(versionTemplate, core.Version, core.GitCommit, core.BuildTime)

	styleGrepMatch = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
)

type entryInfo struct {
	name     string
	size     uint64
	modified time.Time
	isDir    bool
}

func main() {
	runApp(os.Args)
}

func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

func mainAction(c *cli.Context) error {
	if len(c.Args()) < 1 {
		cli.ShowAppHelp(c)
		fmt.Fprintf(stderr, "Hint:  %s\n", core.Hint("missing_arg"))
		return cli.NewExitError(missingParamInputPath, 3)
	}
	grepColor = c.String("grep")

	fileName := c.Args()[0]

	if !files.Exists(fileName) {
		fmt.Fprintf(stderr, "Error: file %q not found\n", fileName)
		fmt.Fprintf(stderr, "Hint:  %s\n", core.Hint("file_not_found"))
		cli.ShowAppHelp(c)
		return cli.NewExitError(fmt.Sprintf("Invalid file: %s", fileName), 2)
	}

	var entries []entryInfo
	err := zipext.Walk(fileName, func(f *zip.File, err error) error {
		if err != nil {
			fmt.Fprintf(stderr, "Error reading entry: %s\n", err)
			return err
		}
		if entryRejected(c, f) {
			return nil
		}
		entries = append(entries, entryInfo{
			name:     f.Name,
			size:     f.UncompressedSize64,
			modified: f.Modified,
			isDir:    isDirectory(f),
		})
		return nil
	})

	if err != nil {
		fmt.Fprintf(stderr, "Hint:  %s\n", core.Hint("invalid_zip"))
		return cli.NewExitError(fmt.Sprintf("error processing %s: %s", fileName, err.Error()), 1)
	}

	sortEntries(entries, c.String("sort"), c.Bool("reverse"))
	printTable(entries, c.Bool("size"), c.Bool("time"))
	return nil
}

func sortEntries(entries []entryInfo, by string, reverse bool) {
	var less func(i, j int) bool
	switch by {
	case "name":
		less = func(i, j int) bool { return entries[i].name < entries[j].name }
	case "size":
		less = func(i, j int) bool { return entries[i].size > entries[j].size }
	case "date":
		less = func(i, j int) bool { return entries[i].modified.After(entries[j].modified) }
	default:
		return
	}
	if reverse {
		orig := less
		less = func(i, j int) bool { return orig(j, i) }
	}
	sort.SliceStable(entries, less)
}

func printTable(entries []entryInfo, showSize, showTime bool) {
	if len(entries) == 0 {
		return
	}

	nameWidth := len("NAME")
	for _, e := range entries {
		if len(e.name) > nameWidth {
			nameWidth = len(e.name)
		}
	}

	// Header
	header := fmt.Sprintf("%-*s", nameWidth, "NAME")
	if showSize {
		header += "  " + fmt.Sprintf("%-10s", "SIZE")
	}
	if showTime {
		header += "  " + fmt.Sprintf("%-16s", "MODIFIED")
	}
	header += "  TYPE"
	fmt.Fprintln(stdout, header)
	fmt.Fprintln(stdout, strings.Repeat("-", len(header)))

	// Rows: pad based on raw name length so color codes don't break alignment
	for _, e := range entries {
		display := coloredName(e.name)
		padding := strings.Repeat(" ", nameWidth-len(e.name))
		row := display + padding

		if showSize {
			sz := humanSize(e.size)
			if e.isDir {
				sz = "-"
			}
			row += "  " + fmt.Sprintf("%-10s", sz)
		}
		if showTime {
			row += "  " + fmt.Sprintf("%-16s", e.modified.Format("2006-01-02 15:04"))
		}
		entryType := "file"
		if e.isDir {
			entryType = "dir"
		}
		row += "  " + entryType
		fmt.Fprintln(stdout, row)
	}
}

func coloredName(name string) string {
	if grepColor == "" || !colorEnabled() {
		return name
	}
	return strings.Replace(name, grepColor, styleGrepMatch.Render(grepColor), -1)
}

func runApp(args []string) {
	app := cli.NewApp()
	app.Name = "zipls"
	app.Usage = "Read zip files"
	app.Version = appVersion

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "grep, g", Value: "", Usage: "filter entries containing pattern (highlighted in output)"},
		cli.StringFlag{Name: "exclude, e", Value: "", Usage: "exclude entries containing pattern"},
		cli.BoolFlag{Name: "dir, d", Usage: "show only directories"},
		cli.BoolFlag{Name: "size, s", Usage: "show file size"},
		cli.BoolFlag{Name: "time, t", Usage: "show modification time"},
		cli.StringFlag{Name: "sort", Value: "none", Usage: "sort order: none, name, size, date"},
		cli.BoolFlag{Name: "reverse, r", Usage: "reverse sort order"},
	}

	app.Action = mainAction
	app.Run(args)
}

func isDirectory(f *zip.File) bool {
	return strings.HasSuffix(f.Name, "/") && f.UncompressedSize64 == 0
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
	if c.Bool("dir") && !isDirectory(f) {
		return true
	}
	return false
}

func humanSize(size uint64) string {
	i := 0
	sizef := float64(size)
	units := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	for sizef >= 1000.0 {
		sizef /= 1000.0
		i++
	}
	return fmt.Sprintf("%.4g %s", sizef, units[i])
}
