package tui

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/term"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// spinnerFrames is the classic braille dot pattern used by charmbracelet's
// spinner.Dot. Copied verbatim so we don't pull bubbletea into commands that
// only need an inline spinner next to a blocking call.
var spinnerFrames = []string{
	"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
}

// RunWithSpinner shows an animated spinner on stderr while fn runs, then
// clears the line before returning fn's error. When stderr is not a TTY
// (pipes, CI) fn is called straight through with no output, so scripts and
// logs stay clean.
//
// Use this for short-lived blocking calls: registry fetches, manifest loads,
// anything under a few seconds. For multi-step progress, drive a bubbletea
// model with install.EventSink instead.
func RunWithSpinner(theme *ui.Theme, label string, fn func() error) error {
	errW := os.Stderr
	if !term.IsTerminal(int(errW.Fd())) {
		return fn()
	}

	stop := startSpinner(errW, theme, label)
	err := fn()
	stop()
	return err
}

func startSpinner(w io.Writer, theme *ui.Theme, label string) func() {
	done := make(chan struct{})
	var once sync.Once
	stop := func() {
		once.Do(func() {
			close(done)
			// Clear the spinner line before the caller prints anything else.
			fmt.Fprint(w, "\r\x1b[K")
		})
	}

	go func() {
		ticker := time.NewTicker(90 * time.Millisecond)
		defer ticker.Stop()
		i := 0
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				frame := theme.Brand.Render(spinnerFrames[i%len(spinnerFrames)])
				fmt.Fprintf(w, "\r%s %s", frame, theme.Crumb.Render(label))
				i++
			}
		}
	}()
	return stop
}
