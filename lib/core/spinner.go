package core

import (
	"fmt"
	"os"
	"time"

	"github.com/mattn/go-isatty"
)

// StartSpinner displays a spinning indicator to stderr while work is being done.
// It suppresses itself when stderr is not a TTY (e.g. pipes, CI).
// Returns a stop function that clears the spinner line.
func StartSpinner(msg string) func() {
	if !isatty.IsTerminal(os.Stderr.Fd()) && !isatty.IsCygwinTerminal(os.Stderr.Fd()) {
		return func() {}
	}
	frames := []string{"|", "/", "-", "\\"}
	done := make(chan struct{})
	go func() {
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				fmt.Fprintf(os.Stderr, "\r  %s  %s", frames[i%4], msg)
				time.Sleep(80 * time.Millisecond)
				i++
			}
		}
	}()
	return func() {
		close(done)
		fmt.Fprintf(os.Stderr, "\r\033[K")
	}
}
