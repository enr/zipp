package main

import (
	"testing"

	"github.com/enr/zipp/lib/qac"
)

var specs = []qac.ConventionalSpec{
	// {
	// 	CommandExe:  "../../bin/zipw",
	// 	CommandArgs: []string{},
	// 	Success:     false,
	// 	ExitCode:    3,
	// 	Stderr: []string{
	// 		"",
	// 	},
	// },
	// {
	// 	CommandExe:  "../../bin/zipw",
	// 	CommandArgs: []string{"../testdata"},
	// 	WorkingDir:  "../../bin",
	// 	Success:     true,
	// 	ExitCode:    0,
	// 	Stdout:      []string{"Completed"},
	// },
	{
		CommandExe:  "../../bin/zipw",
		CommandArgs: []string{"--version"},
		Success:     true,
		ExitCode:    0,
		Stdout: []string{"zipw version",
			"Revision", "Build date"},
	},
}

func TestCommandExecution2(t *testing.T) {
	guarantor := qac.NewGuarantor()
	for _, spec := range specs {
		result := guarantor.VerifyConventional(spec)
		if len(result.Errors()) > 0 {
			t.Errorf("QAC errors %v", result.Errors())
		}
	}
}
