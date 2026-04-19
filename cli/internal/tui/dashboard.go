package tui

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// DashboardResult is what RunDashboard returns to the launcher loop.
// Command is one of the Tile.Command values (or "" on quit).
type DashboardResult struct {
	Command string
	Quit    bool
}

// DashboardTile is one launchable action in the grid.
type DashboardTile struct {
	Command string   // "install", "list", etc. — surfaced to the launcher
	Label   string   // displayed name
	Hotkey  string   // single-key shortcut (e.g. "i", "/", "R")
	Desc    string   // one-line description
	Sub     string   // muted foot line ("skills · deps · conflicts")
	Status  string   // optional badge text ("3 available", "5 drift", "ok")
	Aliases []string // additional fuzzy-search keywords
}

// DashboardGreeting is the banner displayed above the tile grid.
type DashboardGreeting struct {
	User     string // username ("jennings")
	Updates  int    // how many drifted installs
	Cwd      string // working dir
	LastScan string // "18:04" or ""
}

// DashboardStatus feeds the header's right-anchored summary
// (● healthy · N platforms · M skills).
type DashboardStatus struct {
	Healthy   bool
	Platforms int // detected adapters
	Skills    int // installed skills (unique)
}

// DashboardConfig bundles everything RunDashboard needs.
type DashboardConfig struct {
	Theme    *ui.Theme
	Version  string
	Greeting DashboardGreeting
	Status   DashboardStatus
	Tiles    []DashboardTile
}

// RunDashboard opens the full-screen launcher (bubbletea alt-screen) and
// returns once the user picks a tile or quits.
func RunDashboard(cfg DashboardConfig) (DashboardResult, error) {
	if cfg.Theme == nil {
		cfg.Theme = ui.DefaultTheme()
	}
	m := dashboardModel{
		cfg:      cfg,
		cursor:   0,
		searchOn: false,
	}
	m.rebuildVisible()
	out, err := Run(m)
	if err != nil {
		return DashboardResult{}, err
	}
	fm, ok := out.(dashboardModel)
	if !ok {
		return DashboardResult{}, nil
	}
	return fm.result, nil
}

// DefaultDashboardTiles returns the 9-tile layout from the design handoff.
func DefaultDashboardTiles() []DashboardTile {
	return []DashboardTile{
		{Command: "install", Label: "install", Hotkey: "i", Desc: "add a skill to every detected platform", Sub: "registry → adapters", Aliases: []string{"add", "get"}},
		{Command: "list", Label: "list", Hotkey: "l", Desc: "what's installed, where, and what drifted", Sub: "manifest · adapters", Aliases: []string{"ls", "installed"}},
		{Command: "update", Label: "update", Hotkey: "u", Desc: "pull newer registry versions onto installs", Sub: "diff · apply", Aliases: []string{"upgrade"}},
		{Command: "search", Label: "search", Hotkey: "/", Desc: "browse every skill in the registry", Sub: "fuzzy over name, tag, desc", Aliases: []string{"find", "browse"}},
		{Command: "uninstall", Label: "uninstall", Hotkey: "x", Desc: "remove a skill from every target", Sub: "manifest-aware", Aliases: []string{"remove", "rm", "delete"}},
		{Command: "profile", Label: "profile", Hotkey: "p", Desc: "edit install defaults (platforms, scope)", Sub: "user-wide preferences", Aliases: []string{"config", "prefs"}},
		{Command: "doctor", Label: "doctor", Hotkey: "d", Desc: "inspect adapters and environment health", Sub: "adapters · writability", Aliases: []string{"check", "status"}},
		{Command: "registry", Label: "registry", Hotkey: "R", Desc: "refresh the local registry cache", Sub: "http · etag", Aliases: []string{"refresh", "sync"}},
		{Command: "version", Label: "version", Hotkey: "V", Desc: "show build info", Sub: "version · commit", Aliases: []string{"about", "ver"}},
	}
}

// BuildDashboardGreeting fills the defaults the dashboard banner needs.
func BuildDashboardGreeting(updates int) DashboardGreeting {
	g := DashboardGreeting{Updates: updates}
	if u, err := user.Current(); err == nil && u.Username != "" {
		g.User = u.Username
	}
	if cwd, err := os.Getwd(); err == nil {
		g.Cwd = compactPath(cwd)
	}
	return g
}

func compactPath(p string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	if strings.HasPrefix(p, home) {
		return "~" + strings.TrimPrefix(p, home)
	}
	return p
}

// dashboardModel is the bubbletea model for the launcher.
type dashboardModel struct {
	cfg     DashboardConfig
	cursor  int // index into visible tiles
	visible []int

	// Search
	searchOn bool
	query    string

	width, height int
	done          bool
	result        DashboardResult
}

func (m dashboardModel) Init() tea.Cmd { return nil }

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.searchOn {
			return m.updateSearch(msg)
		}
		return m.updateGrid(msg)
	}
	return m, nil
}

func (m dashboardModel) updateGrid(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	switch k {
	case "ctrl+c", "q", "esc":
		m.result = DashboardResult{Quit: true}
		m.done = true
		return m, tea.Quit
	case "/":
		m.searchOn = true
		m.query = ""
		m.rebuildVisible()
		return m, nil
	case "up", "k":
		m = m.moveCursor(-m.cols())
		return m, nil
	case "down", "j":
		m = m.moveCursor(m.cols())
		return m, nil
	case "left", "h":
		m = m.moveCursor(-1)
		return m, nil
	case "right", "l":
		m = m.moveCursor(1)
		return m, nil
	case "tab":
		m = m.moveCursor(1)
		return m, nil
	case "shift+tab":
		m = m.moveCursor(-1)
		return m, nil
	case "enter":
		return m.launchCursor()
	default:
		// Hotkey letter match.
		if len(k) == 1 {
			for _, t := range m.cfg.Tiles {
				if t.Hotkey == k {
					m.result = DashboardResult{Command: t.Command}
					m.done = true
					return m, tea.Quit
				}
			}
		}
	}
	return m, nil
}

func (m dashboardModel) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	switch k {
	case "ctrl+c":
		m.result = DashboardResult{Quit: true}
		m.done = true
		return m, tea.Quit
	case "esc":
		m.searchOn = false
		m.query = ""
		m.rebuildVisible()
		return m, nil
	case "enter":
		return m.launchCursor()
	case "up", "down", "left", "right", "tab", "shift+tab":
		// Let navigation work while typing.
		return m.updateGrid(msg)
	case "backspace":
		if len(m.query) > 0 {
			m.query = m.query[:len(m.query)-1]
		}
		m.rebuildVisible()
		return m, nil
	default:
		if len(k) == 1 {
			m.query += k
			m.rebuildVisible()
		}
		return m, nil
	}
}

func (m dashboardModel) launchCursor() (tea.Model, tea.Cmd) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return m, nil
	}
	t := m.cfg.Tiles[m.visible[m.cursor]]
	m.result = DashboardResult{Command: t.Command}
	m.done = true
	return m, tea.Quit
}

func (m *dashboardModel) rebuildVisible() {
	q := strings.TrimSpace(m.query)
	if q == "" {
		m.visible = make([]int, len(m.cfg.Tiles))
		for i := range m.cfg.Tiles {
			m.visible[i] = i
		}
		if m.cursor >= len(m.visible) {
			m.cursor = 0
		}
		return
	}
	// Fuzzy over "label alias1 alias2 desc".
	haystack := make([]string, len(m.cfg.Tiles))
	for i, t := range m.cfg.Tiles {
		haystack[i] = strings.ToLower(t.Label + " " + strings.Join(t.Aliases, " ") + " " + t.Desc)
	}
	matches := fuzzy.Find(strings.ToLower(q), haystack)
	vis := make([]int, 0, len(matches))
	for _, mm := range matches {
		vis = append(vis, mm.Index)
	}
	m.visible = vis
	if m.cursor >= len(m.visible) {
		m.cursor = 0
	}
}

func (m dashboardModel) moveCursor(delta int) dashboardModel {
	if len(m.visible) == 0 {
		return m
	}
	next := m.cursor + delta
	if next < 0 {
		next = 0
	}
	if next >= len(m.visible) {
		next = len(m.visible) - 1
	}
	m.cursor = next
	return m
}

func (m dashboardModel) cols() int {
	if m.width >= 120 {
		return 3
	}
	if m.width >= 80 {
		return 2
	}
	return 1
}

func (m dashboardModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	th := m.cfg.Theme
	header := Header(th, HeaderSpec{
		Version: m.cfg.Version,
		Section: "Dashboard",
		Meta:    m.statusLine(),
	}, m.width)

	body := m.renderBody()
	right := th.Meta.Render("focused: ") + th.Brand.Render(m.focusedLabel())
	footer := Footer(th, m.hints(), right, m.width)

	bodyH := m.height - lipgloss.Height(header) - lipgloss.Height(footer) - 1
	if bodyH < 5 {
		bodyH = 5
	}
	return Frame(header, body, footer, bodyH)
}

func (m dashboardModel) focusedLabel() string {
	if m.searchOn {
		return "search"
	}
	if len(m.visible) == 0 {
		return "—"
	}
	return m.cfg.Tiles[m.visible[m.cursor]].Label
}

func (m dashboardModel) hints() []KeyHint {
	if m.searchOn {
		return []KeyHint{
			{Keys: "type", Label: "filter"},
			{Keys: "↑↓←→", Label: "move"},
			{Keys: "↵", Label: "launch"},
			{Keys: "esc", Label: "clear"},
		}
	}
	return []KeyHint{
		{Keys: "↑↓←→", Label: "move"},
		{Keys: "↵", Label: "launch"},
		{Keys: "/", Label: "search"},
		{Keys: "esc", Label: "quit"},
	}
}

// statusLine renders the right-anchored header summary:
// "● healthy · 2 platforms · 7 skills".
func (m dashboardModel) statusLine() string {
	th := m.cfg.Theme
	dot := th.DotOK
	label := "healthy"
	if !m.cfg.Status.Healthy {
		dot = th.DotWarn
		label = "drift"
	}
	sep := th.Crumb.Render(" · ")
	return dot.Render("●") + " " + th.Detail.Render(label) +
		sep + th.Detail.Render(fmt.Sprintf("%d platform%s", m.cfg.Status.Platforms, pluralDash(m.cfg.Status.Platforms))) +
		sep + th.Detail.Render(fmt.Sprintf("%d skill%s", m.cfg.Status.Skills, pluralDash(m.cfg.Status.Skills)))
}

func (m dashboardModel) renderBody() string {
	th := m.cfg.Theme
	var sb strings.Builder
	sb.WriteString(m.renderBanner())
	sb.WriteString("\n\n")
	sb.WriteString(m.renderSearchBar())
	sb.WriteString("\n\n")
	sb.WriteString(m.renderGrid())
	if len(m.visible) == 0 {
		sb.WriteString("\n  " + th.Crumb.Render("no command matches "+fmt.Sprintf("%q", m.query)))
	}
	return sb.String()
}

// bodyWidth is the usable width between left/right margins. Every body
// element (banner, search bar, tile grid row) targets this width so they
// all end at the same column.
func (m dashboardModel) bodyWidth() int {
	w := m.width - 4
	if w < 40 {
		w = 40
	}
	return w
}

func (m dashboardModel) renderBanner() string {
	th := m.cfg.Theme
	wordmark := th.Brand.Render("humblskills")
	greet := "hello"
	if m.cfg.Greeting.User != "" {
		greet = "hi " + m.cfg.Greeting.User
	}
	parts := []string{th.Name.Render(greet)}
	if m.cfg.Greeting.Updates > 0 {
		parts = append(parts, th.Warn.Render(fmt.Sprintf("%d update%s available", m.cfg.Greeting.Updates, pluralDash(m.cfg.Greeting.Updates))))
	} else {
		parts = append(parts, th.Detail.Render("all skills up-to-date"))
	}
	if m.cfg.Greeting.Cwd != "" {
		parts = append(parts, th.Detail.Render("in "+m.cfg.Greeting.Cwd))
	}
	line := strings.Join(parts, th.Crumb.Render("  ·  "))
	return "  " + wordmark + "  " + line
}

func (m dashboardModel) renderSearchBar() string {
	th := m.cfg.Theme
	total := m.bodyWidth() // final display width including border
	// Inner content width = total - 2 (border) - 2 (padding) = total - 4.
	inner := total - 4
	if inner < 10 {
		inner = 10
	}
	sigil := th.Brand.Render("❯")
	var query string
	if m.query == "" && !m.searchOn {
		query = th.Crumb.Render("press / to search")
	} else if m.query == "" {
		query = th.Crumb.Render("type to filter…")
	} else {
		query = th.Name.Render(m.query)
	}
	count := th.Crumb.Render(fmt.Sprintf("%d / %d", len(m.visible), len(m.cfg.Tiles)))
	left := sigil + "  " + query
	line := padBetween(left, count, inner)
	borderColor := th.Palette.Border
	if m.searchOn {
		borderColor = th.Palette.Magenta
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(inner).
		Render(line)
	return "  " + box
}

func (m dashboardModel) renderGrid() string {
	if len(m.visible) == 0 {
		return ""
	}
	cols := m.cols()
	gap := 2
	body := m.bodyWidth()
	tileW := (body - (cols-1)*gap) / cols
	if tileW < 24 {
		tileW = 24
	}
	spacer := strings.Repeat(" ", gap)

	rows := [][]string{}
	var row []string
	for i, idx := range m.visible {
		tile := m.renderTile(m.cfg.Tiles[idx], i == m.cursor, tileW)
		row = append(row, tile)
		if len(row) == cols {
			rows = append(rows, row)
			row = nil
		}
	}
	if len(row) > 0 {
		empty := m.renderEmptyTile(tileW)
		for len(row) < cols {
			row = append(row, empty)
		}
		rows = append(rows, row)
	}

	var sb strings.Builder
	for i, r := range rows {
		parts := make([]string, 0, len(r)*2-1)
		for j, t := range r {
			if j > 0 {
				parts = append(parts, spacer)
			}
			parts = append(parts, t)
		}
		sb.WriteString("  ")
		sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, parts...))
		if i < len(rows)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// renderTile renders one tile. The returned string is exactly `width`
// display columns wide (border included) so JoinHorizontal aligns columns.
func (m dashboardModel) renderTile(t DashboardTile, selected bool, width int) string {
	th := m.cfg.Theme
	borderColor := th.Palette.Border
	if selected {
		borderColor = th.Palette.Magenta
	}
	nameStyle := th.RowUnselected
	if selected {
		nameStyle = th.RowSelected
	}

	// Inner width: total minus 2 (border) minus 2 (padding).
	inner := width - 4
	if inner < 10 {
		inner = 10
	}

	hot := th.KbdKey.Render(t.Hotkey)
	name := nameStyle.Render(t.Label)
	header := padBetween(name, hot, inner)

	desc := th.Desc.Width(inner).Render(t.Desc)

	footLeft := th.Detail.Render(t.Sub)
	footRight := ""
	if t.Status != "" {
		footRight = th.BadgeGhost.Render(t.Status)
	}
	foot := padBetween(footLeft, footRight, inner)

	body := header + "\n\n" + desc + "\n\n" + foot
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(width - 2).
		Render(body)
}

// renderEmptyTile returns a blank block the same height as a real tile so
// short final rows still align under a full row above.
func (m dashboardModel) renderEmptyTile(width int) string {
	return lipgloss.NewStyle().
		Width(width).
		Height(5).
		Render("")
}

func pluralDash(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
