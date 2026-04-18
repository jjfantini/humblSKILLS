// Package ui is the single place humblskills prints user-facing text.
// Centralising output here keeps --no-color, --quiet, --verbose, --json,
// and the NO_COLOR env var honest across every subcommand.
package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Level controls how much Printer prints.
type Level int

const (
	// LevelQuiet prints only Error.
	LevelQuiet Level = iota - 1
	// LevelNormal prints Info, Success, Warn, Error.
	LevelNormal
	// LevelVerbose additionally prints Detail.
	LevelVerbose
)

// Options configures a Printer. Zero-value is safe: writes to stdout/stderr,
// normal level, colour auto-detected.
type Options struct {
	Out     io.Writer
	Err     io.Writer
	Level   Level
	NoColor bool
	// JSON suppresses human-readable output. Errors still go to Err so
	// failures aren't silent, but Info/Success/Warn/Detail become no-ops
	// so a --json consumer only sees the JSON document.
	JSON bool
}

// Printer writes styled messages according to its configured options.
type Printer struct {
	out, err io.Writer
	level    Level
	jsonMode bool
	theme    *Theme
}

// New returns a Printer configured by opts. It honours the NO_COLOR env var in
// addition to opts.NoColor.
func New(opts Options) *Printer {
	out := opts.Out
	if out == nil {
		out = os.Stdout
	}
	errW := opts.Err
	if errW == nil {
		errW = os.Stderr
	}

	noColor := opts.NoColor || os.Getenv("NO_COLOR") != ""

	return &Printer{
		out:      out,
		err:      errW,
		level:    opts.Level,
		jsonMode: opts.JSON,
		theme:    NewTheme(DefaultPalette(), out, noColor),
	}
}

// Theme returns the Printer's underlying theme, so commands rendering ad-hoc
// lipgloss layouts (search, doctor, TUIs) share the same palette + renderer
// and honour --no-color without reaching for their own state.
func (p *Printer) Theme() *Theme { return p.theme }

// Println writes a pre-rendered string to stdout with a trailing newline.
// Use when the caller has already composed styled output (e.g. via lipgloss)
// and just needs it emitted. Suppressed in JSON / quiet mode.
func (p *Printer) Println(s string) {
	if p.jsonMode || p.level < LevelNormal {
		return
	}
	fmt.Fprintln(p.out, s)
}

// Info prints at LevelNormal+. Suppressed in JSON mode.
func (p *Printer) Info(format string, args ...any) {
	if p.jsonMode || p.level < LevelNormal {
		return
	}
	fmt.Fprintln(p.out, p.theme.Info.Render(fmt.Sprintf(format, args...)))
}

// Success prints a styled success line. Suppressed in JSON / quiet mode.
func (p *Printer) Success(format string, args ...any) {
	if p.jsonMode || p.level < LevelNormal {
		return
	}
	fmt.Fprintln(p.out, p.theme.Success.Render("✓ ")+fmt.Sprintf(format, args...))
}

// Warn prints a styled warning. Suppressed in JSON / quiet mode.
func (p *Printer) Warn(format string, args ...any) {
	if p.jsonMode || p.level < LevelNormal {
		return
	}
	fmt.Fprintln(p.err, p.theme.Warn.Render("! ")+fmt.Sprintf(format, args...))
}

// Error prints a styled error. Always emitted (even in quiet / JSON mode) so
// failures aren't silent.
func (p *Printer) Error(format string, args ...any) {
	fmt.Fprintln(p.err, p.theme.Error.Render("✗ ")+fmt.Sprintf(format, args...))
}

// Detail prints at LevelVerbose. Suppressed otherwise.
func (p *Printer) Detail(format string, args ...any) {
	if p.jsonMode || p.level < LevelVerbose {
		return
	}
	fmt.Fprintln(p.out, p.theme.Detail.Render(fmt.Sprintf(format, args...)))
}

// JSON marshals v and prints it to Out. Intended for --json output.
func (p *Printer) JSON(v any) error {
	enc := json.NewEncoder(p.out)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// IsJSON reports whether the printer is in JSON mode.
func (p *Printer) IsJSON() bool { return p.jsonMode }

// Out / Err expose the configured writers so commands that compose their own
// lipgloss layouts can emit directly while the Printer still governs colour +
// level semantics.
func (p *Printer) Out() io.Writer { return p.out }
func (p *Printer) Err() io.Writer { return p.err }

// Header prints the shared breadcrumb used above command output:
//
//	  humblskills › <command>
//	  ────────────────────────
//
// Suppressed in JSON / quiet mode.
func (p *Printer) Header(command string) {
	if p.jsonMode || p.level < LevelNormal {
		return
	}
	line := "  " + p.theme.Brand.Render("humblskills") +
		p.theme.Crumb.Render("  ›  "+command)
	fmt.Fprintln(p.out)
	fmt.Fprintln(p.out, line)
}

// Section prints a muted label used to group related lines in static output
// (e.g. "Adapters:", "Registry:"). Suppressed in JSON / quiet mode.
func (p *Printer) Section(label string) {
	if p.jsonMode || p.level < LevelNormal {
		return
	}
	fmt.Fprintln(p.out)
	fmt.Fprintln(p.out, p.theme.Label.Render(label))
}
