package core

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
	reader    io.Reader
}

// NewUI returns a UI writing to os.Stdout/Stderr at the given verbosity level.
func NewUI(verbosity int) *UI {
	return &UI{
		Verbosity: verbosity,
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		reader:    os.Stdin,
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

// QuestionWithDefault prints a prompt and reads a line from stdin.
// Returns defaultAnswer if the user presses enter without typing anything.
func (u *UI) QuestionWithDefault(question, defaultAnswer string) (string, error) {
	fmt.Fprintf(u.stdout, "%s [%s]: ", question, defaultAnswer)
	r := bufio.NewReader(u.reader)
	answer, err := r.ReadString('\n')
	if err != nil {
		return defaultAnswer, err
	}
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return defaultAnswer, nil
	}
	return answer, nil
}
