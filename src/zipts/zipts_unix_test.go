// +build darwin freebsd linux netbsd openbsd

package main

import (
	"testing"
)

// linux ok
var validInputPaths = []inputs{
	{"", getCurrentDirectory()},
	{"   ", getCurrentDirectory()},
	{".", getCurrentDirectory()},
	{"testdata", getCurrentDirectory() + "/testdata"},
	{"/input/path", "/input/path"},
	{"/input/path/", "/input/path"},
}

func TestResolveInputPath(t *testing.T) {
	for _, data := range validInputPaths {
		actual, err := resolveInputPath(data.arg)
		if err != nil {
			t.Errorf(`arg="%s" unexpected error %v`, data.arg, err)
		}
		if actual != data.res {
			t.Errorf(`arg="%s" got="%s" expected="%s"`, data.arg, actual, data.res)
		}
	}
}
