package ui

import (
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Palette is the set of semantic colour tokens used across humblskills. One
// palette drives both the Printer's message styles and every lipgloss/bubbletea
// surface so the CLI looks like a single cohesive application.
type Palette struct {
	Brand    lipgloss.Color // primary accent — wordmark, bullets, progress fill
	Accent   lipgloss.Color // secondary accent — platform chips
	Name     lipgloss.Color // skill names, headings
	Tag      lipgloss.Color // #tags
	Platform lipgloss.Color // platform chips
	Hit      lipgloss.Color // filter / query match highlight
	Muted    lipgloss.Color // secondary text, crumbs, versions
	Rule     lipgloss.Color // divider lines
	Desc     lipgloss.Color // long-form description text
	Success  lipgloss.Color
	Warn     lipgloss.Color
	Error    lipgloss.Color
}

// DefaultPalette mirrors the Catppuccin-aligned tokens previously scattered
// across search.go and ui.go. Truecolour hexes degrade gracefully on 256-colour
// terminals via lipgloss.
func DefaultPalette() Palette {
	return Palette{
		Brand:    lipgloss.Color("#A78BFA"),
		Accent:   lipgloss.Color("#F0ABFC"),
		Name:     lipgloss.Color("#5EEAD4"),
		Tag:      lipgloss.Color("#93C5FD"),
		Platform: lipgloss.Color("#F0ABFC"),
		Hit:      lipgloss.Color("#FDE68A"),
		Muted:    lipgloss.Color("244"),
		Rule:     lipgloss.Color("238"),
		Desc:     lipgloss.Color("252"),
		Success:  lipgloss.Color("10"),
		Warn:     lipgloss.Color("11"),
		Error:    lipgloss.Color("9"),
	}
}

// Theme pairs a palette with pre-built lipgloss styles, all tied to a single
// renderer so that --no-color / NO_COLOR degrade in exactly one place.
type Theme struct {
	Palette  Palette
	Renderer *lipgloss.Renderer

	// Printer-facing styles.
	Info    lipgloss.Style
	Success lipgloss.Style
	Warn    lipgloss.Style
	Error   lipgloss.Style
	Detail  lipgloss.Style

	// Layout / content styles consumed by search, doctor, and the tui package.
	Brand    lipgloss.Style
	Crumb    lipgloss.Style
	RuleLine lipgloss.Style
	Bullet   lipgloss.Style
	Name     lipgloss.Style
	Version  lipgloss.Style
	Desc     lipgloss.Style
	Tag      lipgloss.Style
	Platform lipgloss.Style
	Hit      lipgloss.Style
	Label    lipgloss.Style
	Count    lipgloss.Style
	Panel    lipgloss.Style
}

// NewTheme builds a Theme using a renderer attached to out. When noColor is
// true the renderer is coerced to ASCII so every style becomes a no-op.
func NewTheme(palette Palette, out io.Writer, noColor bool) *Theme {
	if out == nil {
		out = os.Stdout
	}
	var r *lipgloss.Renderer
	if noColor {
		r = lipgloss.NewRenderer(out)
		r.SetColorProfile(termenv.Ascii)
	} else {
		r = lipgloss.DefaultRenderer()
	}
	return buildTheme(palette, r)
}

// DefaultTheme returns the standard palette wrapping the default lipgloss
// renderer. Suitable for commands that don't need to override colour handling.
func DefaultTheme() *Theme {
	return buildTheme(DefaultPalette(), lipgloss.DefaultRenderer())
}

func buildTheme(p Palette, r *lipgloss.Renderer) *Theme {
	t := &Theme{Palette: p, Renderer: r}

	t.Info = r.NewStyle()
	t.Success = r.NewStyle().Foreground(p.Success).Bold(true)
	t.Warn = r.NewStyle().Foreground(p.Warn).Bold(true)
	t.Error = r.NewStyle().Foreground(p.Error).Bold(true)
	t.Detail = r.NewStyle().Foreground(p.Muted)

	t.Brand = r.NewStyle().Bold(true).Foreground(p.Brand)
	t.Crumb = r.NewStyle().Foreground(p.Muted)
	t.RuleLine = r.NewStyle().Foreground(p.Rule)
	t.Bullet = r.NewStyle().Foreground(p.Brand)
	t.Name = r.NewStyle().Bold(true).Foreground(p.Name)
	t.Version = r.NewStyle().Foreground(p.Muted)
	t.Desc = r.NewStyle().Foreground(p.Desc)
	t.Tag = r.NewStyle().Foreground(p.Tag)
	t.Platform = r.NewStyle().Foreground(p.Platform)
	t.Hit = r.NewStyle().Bold(true).Foreground(p.Hit)
	t.Label = r.NewStyle().Foreground(p.Muted)
	t.Count = r.NewStyle().Foreground(p.Muted).Italic(true)
	t.Panel = r.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Rule).
		Padding(0, 1)

	return t
}
