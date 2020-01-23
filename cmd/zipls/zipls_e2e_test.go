package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/enr/runcmd"
)

type commandExecution struct {
	command  *runcmd.Command
	success  bool
	exitCode int
	stdout   string
	stderr   string
}

func exePath(p string) string {
	adjusted := fmt.Sprintf("../../bin/%s", p)
	a, _ := filepath.Abs(adjusted)
	var ext string
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	executablePath := fmt.Sprintf("%s%s", a, ext)
	if _, err := os.Stat(executablePath); os.IsNotExist(err) {
		panic(fmt.Sprintf("no such executable: %s", executablePath))
	}
	return executablePath
}

var executions = []commandExecution{
	{
		command: &runcmd.Command{
			Exe:  exePath("zipls"),
			Args: []string{},
		},
		success:  false,
		exitCode: 3,
		stderr:   missingParamInputPath,
	},
	{
		command: &runcmd.Command{
			Exe:  exePath("zipls"),
			Args: []string{"--version"},
		},
		success:  true,
		exitCode: 0,
		stdout:   "zipls version",
		//  fmt.Sprintf("zipls version %s", appVersion),
	},
}

func TestCommandExecution(t *testing.T) {
	for _, d := range executions {
		command := d.command
		res := command.Run()
		if res.Success() != d.success {
			t.Fatalf("%s: expected success %t but got %t", command, d.success, res.Success())
		}
		expectedCode := d.exitCode
		actualCode := res.ExitStatus()
		if actualCode != expectedCode {
			t.Fatalf("%s: expected exit code %d but got %d", command, expectedCode, actualCode)
		}
		assertStringContains(t, res.Stdout().String(), d.stdout)
		assertStringContains(t, res.Stderr().String(), d.stderr)
	}
}

func assertStringContains(t *testing.T, s string, substr string) {
	if substr != "" && !strings.Contains(s, substr) {
		t.Fatalf("expected output\n%s\n  does not contain\n%s\n", s, substr)
	}
}
