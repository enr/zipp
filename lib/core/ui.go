package core

import (
	"fmt"
	"io"
	"os"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/mattn/go-isatty"
)

// Verbosity levels for UI output.
const (
	VerbosityQuiet   = 0
	VerbosityMedium  = 1
	VerbosityVerbose = 2
)

var (
	styleError     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleSuccess   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleLifecycle = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	styleDim       = lipgloss.NewStyle().Faint(true)
)

// UI handles terminal output with verbosity levels and optional color.
type UI struct {
	Verbosity int
	stdout    io.Writer
	stderr    io.Writer
}

// NewUI returns a UI writing to os.Stdout/Stderr at the given verbosity level.
func NewUI(verbosity int) *UI {
	return &UI{
		Verbosity: verbosity,
		stdout:    os.Stdout,
		stderr:    os.Stderr,
	}
}

func (u *UI) Error(msg string) {
	fmt.Fprintln(u.stderr, styleError.Render(msg))
}

// Errorf formats and prints an error message to stderr.
func (u *UI) Errorf(format string, args ...interface{}) {
	u.Error(fmt.Sprintf(format, args...))
}

// Successf formats and prints a success message when verbosity >= VerbosityMedium.
func (u *UI) Successf(format string, args ...interface{}) {
	if u.Verbosity >= VerbosityMedium {
		fmt.Fprintln(u.stdout, styleSuccess.Render(fmt.Sprintf(format, args...)))
	}
}

// Lifecyclef formats and prints a lifecycle message when verbosity >= VerbosityMedium.
func (u *UI) Lifecyclef(format string, args ...interface{}) {
	if u.Verbosity >= VerbosityMedium {
		fmt.Fprintln(u.stdout, styleLifecycle.Render(fmt.Sprintf(format, args...)))
	}
}

// Confidentialf formats and prints a dimmed message to stderr when verbosity >= VerbosityVerbose.
func (u *UI) Confidentialf(format string, args ...interface{}) {
	if u.Verbosity >= VerbosityVerbose {
		fmt.Fprintln(u.stderr, styleDim.Render(fmt.Sprintf(format, args...)))
	}
}

func isInteractive() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

func newForm(groups ...*huh.Group) *huh.Form {
	f := huh.NewForm(groups...)
	if !isInteractive() {
		f = f.WithAccessible(true)
	}
	return f
}

// Confirm shows an interactive yes/no prompt. Returns the boolean result.
func (u *UI) Confirm(title string, defaultValue bool) (bool, error) {
	var result bool = defaultValue
	err := newForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Value(&result),
		),
	).Run()
	if err != nil {
		return defaultValue, err
	}
	return result, nil
}

// Input shows an interactive text input prompt. Returns the entered value or defaultValue if empty.
func (u *UI) Input(title, placeholder, defaultValue string) (string, error) {
	result := defaultValue
	err := newForm(
		huh.NewGroup(
			huh.NewInput().
				Title(title).
				Placeholder(placeholder).
				Value(&result),
		),
	).Run()
	if err != nil {
		return defaultValue, err
	}
	if result == "" {
		return defaultValue, nil
	}
	return result, nil
}

// Select shows an interactive selection prompt. Returns the selected value.
func (u *UI) Select(title string, options []string) (string, error) {
	var result string
	opts := make([]huh.Option[string], len(options))
	for i, o := range options {
		opts[i] = huh.NewOption(o, o)
	}
	err := newForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(opts...).
				Value(&result),
		),
	).Run()
	return result, err
}
