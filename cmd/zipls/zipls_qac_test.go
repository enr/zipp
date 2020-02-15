package main

import (
	"testing"

	"github.com/enr/zipp/lib/qac"
)

// type commandExecution struct {
// 	command  *runcmd.Command
// 	success  bool
// 	exitCode int
// 	stdout   string
// 	stderr   string
// }

// func exePath(p string) string {
// 	adjusted := fmt.Sprintf("../../bin/%s", p)
// 	a, _ := filepath.Abs(adjusted)
// 	var ext string
// 	if runtime.GOOS == "windows" {
// 		ext = ".exe"
// 	}
// 	executablePath := fmt.Sprintf("%s%s", a, ext)
// 	if _, err := os.Stat(executablePath); os.IsNotExist(err) {
// 		panic(fmt.Sprintf("no such executable: %s", executablePath))
// 	}
// 	return executablePath
// }

var specs = []qac.Spec{
	{
		Command: qac.Command{
			Exe:  qac.FullExePathOrFail("../../bin/zipls"),
			Args: []string{},
		},
		Expectation: qac.Expectation{
			Status: qac.StatusExpectation{
				Success: false,
				Code:    3,
			},
			Output: qac.OutputExpectations{
				Stderr: qac.OutputExpectation{
					Tokens: []string{
						missingParamInputPath,
					},
					Comparison: qac.Exact,
				},
			},
		},
	},
}

func TestCommandExecution2(t *testing.T) {
	guarantor := qac.NewGuarantor()
	for _, spec := range specs {
		result := guarantor.Verify(spec)
		if len(result.Errors()) > 0 {
			t.Errorf("QAC errors %v", result.Errors())
		}
		// command := d.command
		// res := command.Run()
		// if res.Success() != d.success {
		// 	t.Fatalf("%s: expected success %t but got %t", command, d.success, res.Success())
		// }
		// expectedCode := d.exitCode
		// actualCode := res.ExitStatus()
		// if actualCode != expectedCode {
		// 	t.Fatalf("%s: expected exit code %d but got %d", command, expectedCode, actualCode)
		// }
		// assertStringContains(t, res.Stdout().String(), d.stdout)
		// assertStringContains(t, res.Stderr().String(), d.stderr)
	}
}
