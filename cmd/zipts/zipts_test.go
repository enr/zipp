package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/enr/clui"
)

func init() {
	if ui == nil {
		ui = getUI(clui.VerbosityLevelMute)
	}
}

// used in OS specific tests
type inputs struct {
	arg string
	res string
}

// current working directory with "/" as file separator
func getCurrentDirectory() string {
	wd, _ := os.Getwd()
	return filepath.FromSlash(wd)
}

func TestRun(t *testing.T) {
	runConfig := runConfig{
		Args: []string{"notexists"},
		Noop: false,
	}
	showHelp := func() {
		return
	}
	actual := run(runConfig, showHelp)
	if actual == 0 {
		t.Errorf("run(%v) got 0, expected error", runConfig.Args)
	}
}

func TestRunNoop(t *testing.T) {
	runConfig := runConfig{
		Args: []string{"notexists"},
		Noop: true,
	}
	showHelp := func() {
		return
	}
	actual := run(runConfig, showHelp)
	if actual != 0 {
		t.Errorf("run(%v) noop, got %d, expected 0", runConfig.Args, actual)
	}
}
