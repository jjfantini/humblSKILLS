package ui

import (
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Palette is the humblskills colour vocabulary. It's the Tokyo Night (night
// variant) design-handoff palette, with a second layer of "semantic" aliases
// retained so older surfaces that reach for Brand/Name/Accent keep working
// without churn.
type Palette struct {
	// --- Tokyo Night raw tokens (hex from tui/project/tui.html).
	BG      lipgloss.Color // terminal background (the terminal draws it)
	BGDark  lipgloss.Color // dark text on coloured badges
	BGHL    lipgloss.Color // selected-row bg, kbd chip bg
	Border  lipgloss.Color // dashed rules, 1-col divider
	FG      lipgloss.Color // primary text
	FGDim   lipgloss.Color // secondary text
	Comment lipgloss.Color // labels, crumbs, version, section titles
	Blue    lipgloss.Color // rw badge
	Cyan    lipgloss.Color // path values, scope labels, highlights
	Green   lipgloss.Color // ok dot, detected badge, success
	Magenta lipgloss.Color // brand, selected-row accent
	Red     lipgloss.Color // no dot, missing badge, error
	Yellow  lipgloss.Color // ro badge, warning
	Orange  lipgloss.Color // reserved
	Teal    lipgloss.Color // reserved

	// --- Semantic aliases (kept so old call sites still compile).
	Brand    lipgloss.Color // wordmark, bullets — Magenta
	Accent   lipgloss.Color // secondary accent — Cyan
	Name     lipgloss.Color // names, headings — FG
	Tag      lipgloss.Color // #tags — Blue
	Platform lipgloss.Color // platform chips — Cyan
	Hit      lipgloss.Color // filter match highlight — Yellow
	Muted    lipgloss.Color // crumbs — Comment
	Rule     lipgloss.Color // dividers — Border
	Desc     lipgloss.Color // long-form text — FGDim
	Success  lipgloss.Color
	Warn     lipgloss.Color
	Error    lipgloss.Color
}

// DefaultPalette is the Tokyo Night (night) palette verbatim from the design
// handoff. Hex values degrade gracefully on 256-colour terminals and collapse
// to ASCII under NO_COLOR via lipgloss + termenv.
func DefaultPalette() Palette {
	p := Palette{
		BG:      lipgloss.Color("#1a1b26"),
		BGDark:  lipgloss.Color("#16161e"),
		BGHL:    lipgloss.Color("#292e42"),
		Border:  lipgloss.Color("#3b4261"),
		FG:      lipgloss.Color("#c0caf5"),
		FGDim:   lipgloss.Color("#a9b1d6"),
		Comment: lipgloss.Color("#565f89"),
		Blue:    lipgloss.Color("#7aa2f7"),
		Cyan:    lipgloss.Color("#7dcfff"),
		Green:   lipgloss.Color("#9ece6a"),
		Magenta: lipgloss.Color("#bb9af7"),
		Red:     lipgloss.Color("#f7768e"),
		Yellow:  lipgloss.Color("#e0af68"),
		Orange:  lipgloss.Color("#ff9e64"),
		Teal:    lipgloss.Color("#73daca"),
	}
	p.Brand = p.Magenta
	p.Accent = p.Cyan
	p.Name = p.FG
	p.Tag = p.Blue
	p.Platform = p.Cyan
	p.Hit = p.Yellow
	p.Muted = p.Comment
	p.Rule = p.Border
	p.Desc = p.FGDim
	p.Success = p.Green
	p.Warn = p.Yellow
	p.Error = p.Red
	return p
}

// Theme pairs a palette with pre-built lipgloss styles. All styles share one
// renderer so --no-color / NO_COLOR degrade exactly once.
type Theme struct {
	Palette  Palette
	Renderer *lipgloss.Renderer

	// Printer-facing.
	Info    lipgloss.Style
	Success lipgloss.Style
	Warn    lipgloss.Style
	Error   lipgloss.Style
	Detail  lipgloss.Style

	// Chrome.
	Brand        lipgloss.Style // magenta bold wordmark
	Version      lipgloss.Style // comment "v0.4.2"
	Crumb        lipgloss.Style // comment "· section"
	Meta         lipgloss.Style // comment right-anchored header meta
	RuleLine     lipgloss.Style // dashed rule colour
	Divider      lipgloss.Style // 1-col vertical divider
	SectionTitle lipgloss.Style // uppercase comment header in each pane

	// Row rendering.
	RowSelected   lipgloss.Style // magenta fg, bold (inherits parent bg)
	RowUnselected lipgloss.Style
	RowDim        lipgloss.Style // comment-colour (missing/unavailable)
	RowBg         lipgloss.Style // bg-hl full-row background for the cursor row
	Bullet        lipgloss.Style // magenta ▌ bar
	DotOK         lipgloss.Style // green ●
	DotNo         lipgloss.Style // red ●
	DotWarn       lipgloss.Style // yellow ●

	// Right-pane / detail.
	DetailTitle lipgloss.Style
	DetailSub   lipgloss.Style
	KVKey       lipgloss.Style
	KVValue     lipgloss.Style
	PathLabel   lipgloss.Style
	PathValue   lipgloss.Style

	// Footer kbd chips.
	KbdKey   lipgloss.Style
	KbdLabel lipgloss.Style

	// Badges (reverse-video pills).
	BadgeDetected lipgloss.Style
	BadgeMissing  lipgloss.Style
	BadgeRW       lipgloss.Style
	BadgeRO       lipgloss.Style
	BadgeGhost    lipgloss.Style

	// --- Legacy surfaces (kept so non-migrated callers still compile).
	Name     lipgloss.Style
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

// DefaultTheme returns the standard palette wrapping the default renderer.
func DefaultTheme() *Theme {
	return buildTheme(DefaultPalette(), lipgloss.DefaultRenderer())
}

func buildTheme(p Palette, r *lipgloss.Renderer) *Theme {
	t := &Theme{Palette: p, Renderer: r}

	// Printer.
	t.Info = r.NewStyle().Foreground(p.FG)
	t.Success = r.NewStyle().Foreground(p.Green).Bold(true)
	t.Warn = r.NewStyle().Foreground(p.Yellow).Bold(true)
	t.Error = r.NewStyle().Foreground(p.Red).Bold(true)
	t.Detail = r.NewStyle().Foreground(p.Comment)

	// Chrome.
	t.Brand = r.NewStyle().Foreground(p.Magenta).Bold(true)
	t.Version = r.NewStyle().Foreground(p.Comment)
	t.Crumb = r.NewStyle().Foreground(p.Comment)
	t.Meta = r.NewStyle().Foreground(p.Comment)
	t.RuleLine = r.NewStyle().Foreground(p.Border)
	t.Divider = r.NewStyle().Foreground(p.Border)
	t.SectionTitle = r.NewStyle().Foreground(p.Comment).Bold(true)

	// Rows. RowSelected is applied to the name text only; RowBg wraps the
	// full row so the highlight fills edge-to-edge through any padding gaps.
	t.RowSelected = r.NewStyle().Foreground(p.Magenta).Bold(true)
	t.RowUnselected = r.NewStyle().Foreground(p.FG)
	t.RowDim = r.NewStyle().Foreground(p.Comment)
	t.RowBg = r.NewStyle().Background(p.BGHL)
	t.Bullet = r.NewStyle().Foreground(p.Magenta).Bold(true)
	// Status dots use fixed vibrant hues + bold so they pop on light
	// terminals. AdaptiveColor was tried first, but background detection
	// is unreliable in many terminals (including Claude Code's embedded
	// one), which left users with the dark-mode pastels on white. Fixed
	// saturated tones render well on both backgrounds.
	t.DotOK = r.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true)
	t.DotNo = r.NewStyle().Foreground(lipgloss.Color("#ef4444")).Bold(true)
	t.DotWarn = r.NewStyle().Foreground(lipgloss.Color("#eab308")).Bold(true)

	// Detail pane.
	t.DetailTitle = r.NewStyle().Foreground(p.FG).Bold(true)
	t.DetailSub = r.NewStyle().Foreground(p.Comment).Italic(true)
	t.KVKey = r.NewStyle().Foreground(p.Comment)
	t.KVValue = r.NewStyle().Foreground(p.Cyan)
	t.PathLabel = r.NewStyle().Foreground(p.Cyan).Bold(true)
	t.PathValue = r.NewStyle().Foreground(p.Cyan)

	// Footer kbd chip: bg-hl background, fg text, bottom-inset border.
	t.KbdKey = r.NewStyle().
		Background(p.BGHL).
		Foreground(p.FG).
		Padding(0, 1).
		Bold(true)
	t.KbdLabel = r.NewStyle().Foreground(p.Comment)

	// Badges. Reverse-video pills (dark fg on coloured bg, bold, 1ch padding).
	badge := func(bg lipgloss.Color) lipgloss.Style {
		return r.NewStyle().
			Background(bg).
			Foreground(p.BGDark).
			Padding(0, 1).
			Bold(true)
	}
	t.BadgeDetected = badge(p.Green)
	t.BadgeMissing = badge(p.Red)
	t.BadgeRW = badge(p.Blue)
	t.BadgeRO = badge(p.Yellow)
	t.BadgeGhost = r.NewStyle().
		Foreground(p.Comment).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, true, false, true).
		BorderForeground(p.Border)

	// Legacy styles (Tokyo Night mapped).
	t.Name = r.NewStyle().Foreground(p.FG).Bold(true)
	t.Desc = r.NewStyle().Foreground(p.FGDim)
	t.Tag = r.NewStyle().Foreground(p.Blue)
	t.Platform = r.NewStyle().Foreground(p.Cyan)
	t.Hit = r.NewStyle().Foreground(p.Yellow).Bold(true)
	t.Label = r.NewStyle().Foreground(p.Comment)
	t.Count = r.NewStyle().Foreground(p.Comment).Italic(true)
	t.Panel = r.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Border).
		Padding(0, 1)

	return t
}
