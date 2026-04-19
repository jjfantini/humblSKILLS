package tui

import (
	"fmt"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// VersionInfo is the payload the version screen renders.
type VersionInfo struct {
	Version string
	Commit  string
	Dirty   bool
}

// VersionScreenConfig bundles every field the version TUI needs — including
// the shared breadcrumb + status line so the header matches other sub-screens
// when the user navigates in from the dashboard.
type VersionScreenConfig struct {
	Theme   *ui.Theme
	Section string // "Version" (direct invocation) or "Dashboard > Version" (from launcher)
	Meta    string // right-anchored meta — typically RenderStatusMeta(theme, status)
	Info    VersionInfo
}

// RunVersionScreen opens the version info as a full TUI page so the user
// stays inside the alt-screen experience instead of dropping back to the
// shell. Any key exits.
func RunVersionScreen(cfg VersionScreenConfig) error {
	if cfg.Theme == nil {
		cfg.Theme = ui.DefaultTheme()
	}
	if cfg.Section == "" {
		cfg.Section = "Version"
	}
	_, err := Run(versionScreenModel{cfg: cfg})
	return err
}

type versionScreenModel struct {
	cfg           VersionScreenConfig
	width, height int
}

func (m versionScreenModel) Init() tea.Cmd { return nil }

func (m versionScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "enter", " ":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m versionScreenModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	th := m.cfg.Theme
	header := Header(th, HeaderSpec{
		Version: m.cfg.Info.Version,
		Section: m.cfg.Section,
		Meta:    m.cfg.Meta,
	}, m.width)

	footer := Footer(th, []KeyHint{
		{Keys: "esc", Label: "back"},
		{Keys: "q", Label: "quit"},
	}, "", m.width)

	body := m.renderBody()
	bodyH := m.height - lipgloss.Height(header) - lipgloss.Height(footer) - 1
	if bodyH < 3 {
		bodyH = 3
	}
	return Frame(header, body, footer, bodyH)
}

func (m versionScreenModel) renderBody() string {
	th := m.cfg.Theme
	width := m.width - 4
	if width < 40 {
		width = 40
	}
	inner := width - 4
	if inner < 20 {
		inner = 20
	}

	wordmark := th.Brand.Render("humblskills")
	ver := th.Version.Render(m.cfg.Info.Version)
	if m.cfg.Info.Dirty {
		ver += " " + th.Warn.Render("(dirty)")
	}

	rows := []struct {
		label string
		value string
	}{
		{"version", m.cfg.Info.Version},
		{"commit", m.cfg.Info.Commit},
		{"go", runtime.Version()},
		{"os/arch", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)},
	}

	var sb strings.Builder
	sb.WriteString(wordmark + "   " + ver)
	if m.cfg.Info.Dirty {
		sb.WriteString("\n" + th.Warn.Render("working tree is dirty — uncommitted changes baked into this build"))
	}
	sb.WriteString("\n\n")
	for _, r := range rows {
		sb.WriteString(th.Crumb.Render(padRight(r.label, 10)) + "  " + th.Name.Render(r.value) + "\n")
	}
	sb.WriteString("\n")
	sb.WriteString(th.Detail.Render("press any key to return"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(th.Palette.Border).
		Padding(1, 2).
		Width(inner).
		Render(sb.String())

	return indentBlock(box, 2)
}

// padRight is a small local helper — duplicated from cmd/humblskills to avoid
// an import cycle. Kept tiny on purpose.
func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}
