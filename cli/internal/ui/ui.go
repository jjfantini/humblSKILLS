// Package ui is the single place humblskills prints user-facing text.
// Centralising output here keeps --no-color, --quiet, --verbose, --json,
// and the NO_COLOR env var honest across every subcommand.
package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
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
	styles   styles
}

type styles struct {
	info    lipgloss.Style
	success lipgloss.Style
	warn    lipgloss.Style
	errS    lipgloss.Style
	detail  lipgloss.Style
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

	r := lipgloss.DefaultRenderer()
	if noColor {
		r = lipgloss.NewRenderer(out)
		r.SetColorProfile(termenv.Ascii)
	}

	s := styles{
		info:    r.NewStyle(),
		success: r.NewStyle().Foreground(lipgloss.Color("10")).Bold(true),
		warn:    r.NewStyle().Foreground(lipgloss.Color("11")).Bold(true),
		errS:    r.NewStyle().Foreground(lipgloss.Color("9")).Bold(true),
		detail:  r.NewStyle().Foreground(lipgloss.Color("8")),
	}

	return &Printer{
		out:      out,
		err:      errW,
		level:    opts.Level,
		jsonMode: opts.JSON,
		styles:   s,
	}
}

// Info prints at LevelNormal+. Suppressed in JSON mode.
func (p *Printer) Info(format string, args ...any) {
	if p.jsonMode || p.level < LevelNormal {
		return
	}
	fmt.Fprintln(p.out, p.styles.info.Render(fmt.Sprintf(format, args...)))
}

// Success prints a styled success line. Suppressed in JSON / quiet mode.
func (p *Printer) Success(format string, args ...any) {
	if p.jsonMode || p.level < LevelNormal {
		return
	}
	fmt.Fprintln(p.out, p.styles.success.Render("✓ ")+fmt.Sprintf(format, args...))
}

// Warn prints a styled warning. Suppressed in JSON / quiet mode.
func (p *Printer) Warn(format string, args ...any) {
	if p.jsonMode || p.level < LevelNormal {
		return
	}
	fmt.Fprintln(p.err, p.styles.warn.Render("! ")+fmt.Sprintf(format, args...))
}

// Error prints a styled error. Always emitted (even in quiet / JSON mode) so
// failures aren't silent.
func (p *Printer) Error(format string, args ...any) {
	fmt.Fprintln(p.err, p.styles.errS.Render("✗ ")+fmt.Sprintf(format, args...))
}

// Detail prints at LevelVerbose. Suppressed otherwise.
func (p *Printer) Detail(format string, args ...any) {
	if p.jsonMode || p.level < LevelVerbose {
		return
	}
	fmt.Fprintln(p.out, p.styles.detail.Render(fmt.Sprintf(format, args...)))
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
