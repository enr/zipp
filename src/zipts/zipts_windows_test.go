// +build windows

package main

import (
	"testing"
)

var validInputPathsWindows = []inputs{
	{"", getCurrentDirectory()},
	{"   ", getCurrentDirectory()},
	{".", getCurrentDirectory()},
	{"testdata", getCurrentDirectory() + `\testdata`},
	{`c:\input\path`, `c:\input\path`},
	{`c:\input\path\`, `c:\input\path`},
	{`C:/input/path/`, `C:\input\path`},
}

func TestResolveInputPathWindows(t *testing.T) {
	for _, data := range validInputPathsWindows {
		actual, err := resolveInputPath(data.arg)
		if err != nil {
			t.Errorf(`arg="%s" unexpected error %v`, data.arg, err)
		}
		if actual != data.res {
			t.Errorf(`arg="%s" got="%s" expected="%s"`, data.arg, actual, data.res)
		}
	}
}
