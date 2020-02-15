package qac

import (
	"fmt"
	"strings"
	"testing"
)

type fixedValueExecutor struct {
	success  bool
	exitCode int
	stdout   string
	stderr   string
}

func (e *fixedValueExecutor) execute(c Command) executionResult {
	return executionResult{
		success:  e.success,
		exitCode: e.exitCode,
		stdout:   e.stdout,
		stderr:   e.stderr,
	}
}

func Test1(t *testing.T) {
	e := &fixedValueExecutor{
		success:  true,
		exitCode: 0,
		stdout:   "stdout",
		stderr:   "stderr",
	}
	sut := newGuarantor(e)

	var spec = Spec{

		Command: Command{
			Exe:  "test",
			Args: []string{},
		},
		Expectation: Expectation{
			Status: StatusExpectation{
				Success: false,
				Code:    6,
			},
			Output: OutputExpectations{
				Stderr: OutputExpectation{
					Tokens: []string{
						"qwerty",
					},
					Comparison: Exact,
				},
			},
		},
	}

	res := sut.Verify(spec)
	for i, err := range res.Errors() {
		fmt.Printf("%d e %v\n", i, err)
	}
	if !atLeastOneErrorContaining(res.Errors(), "qwerty") {
		t.Errorf("Expected at least one error containing <%s>", "qwerty")
	}
	if !atLeastOneErrorContaining(res.Errors(), "expected success") {
		t.Errorf("Expected at least one error containing <%s>", "expected success")
	}
	if !atLeastOneErrorContaining(res.Errors(), "expected exit code") {
		t.Errorf("Expected at least one error containing <%s>", "expected exit code")
	}
	if !atLeastOneErrorContaining(res.Errors(), "actual output") {
		t.Errorf("Expected at least one error containing <%s>", "actual output")
	}
	if len(res.Errors()) != 3 {
		t.Errorf(`errs num %d`, res.Errors())
	}
}

func atLeastOneErrorContaining(errors []error, expected string) bool {
	for _, err := range errors {
		if strings.Contains(err.Error(), expected) {
			return true
		}
	}
	return false
}
