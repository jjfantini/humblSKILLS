package tui

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
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

// DashboardGreeting is retained for API compatibility but is no longer
// rendered in the dashboard body — the top header already shows the status
// line, so a duplicate greeting row was redundant noise.
type DashboardGreeting struct {
	User     string
	Updates  int
	Cwd      string
	LastScan string
}

// DashboardStatus feeds the header's right-anchored summary
// (● healthy · N platforms · M skills). This is also what every sub-screen
// echoes in its own header so the layout is consistent across screens.
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
		{Command: "install", Label: "install", Hotkey: "i", Desc: "add a skill to every detected platform", Sub: "registry → platforms", Aliases: []string{"add", "get"}},
		{Command: "list", Label: "list", Hotkey: "l", Desc: "what's installed, where, and what drifted", Sub: "manifest · platforms", Aliases: []string{"ls", "installed"}},
		{Command: "update", Label: "update", Hotkey: "u", Desc: "pull newer registry versions onto installs", Sub: "diff · apply", Aliases: []string{"upgrade-skills"}},
		{Command: "search", Label: "search", Hotkey: "/", Desc: "browse every skill in the registry", Sub: "fuzzy over name, tag, desc", Aliases: []string{"find", "browse"}},
		{Command: "uninstall", Label: "uninstall", Hotkey: "x", Desc: "remove a skill from every target", Sub: "manifest-aware", Aliases: []string{"remove", "rm", "delete"}},
		{Command: "profile", Label: "profile", Hotkey: "p", Desc: "edit install defaults (platforms, scope)", Sub: "user-wide preferences", Aliases: []string{"config", "prefs"}},
		{Command: "eval", Label: "eval", Hotkey: "e", Desc: "benchmark skills · three-arm · longitudinal", Sub: "runners · trajectories · reports", Aliases: []string{"test", "benchmark", "evaluate"}},
		{Command: "doctor", Label: "doctor", Hotkey: "d", Desc: "inspect platforms and environment health", Sub: "platforms · writability", Aliases: []string{"check", "status"}},
		{Command: "registry", Label: "registry", Hotkey: "R", Desc: "refresh the local registry cache", Sub: "http · etag", Aliases: []string{"refresh", "sync"}},
		{Command: "version", Label: "version", Hotkey: "V", Desc: "show build info", Sub: "version · commit", Aliases: []string{"about", "ver"}},
		{Command: "upgrade", Label: "upgrade", Hotkey: "U", Desc: "upgrade the humblskills CLI itself", Sub: "github releases · checksum verified", Aliases: []string{"self-update"}},
	}
}

// BuildDashboardGreeting fills the defaults the dashboard banner needs.
// Kept for API compatibility — callers may still pass this, but it isn't
// rendered. (See DashboardGreeting.)
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

	vp    viewport.Model
	ready bool

	width, height int
	done          bool
	result        DashboardResult
}

func (m dashboardModel) Init() tea.Cmd { return nil }

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m = m.resizeViewport()
		return m, nil
	case tea.KeyMsg:
		var out tea.Model
		var cmd tea.Cmd
		if m.searchOn {
			out, cmd = m.updateSearch(msg)
		} else {
			out, cmd = m.updateGrid(msg)
		}
		dm, ok := out.(dashboardModel)
		if !ok || dm.done {
			return out, cmd
		}
		dm = dm.syncViewport()
		return dm, cmd
	case tea.MouseMsg:
		if !m.ready {
			return m, nil
		}
		vp, cmd := m.vp.Update(msg)
		m.vp = vp
		return m, cmd
	}
	return m, nil
}

func (m dashboardModel) updateGrid(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Shared motions come from the canonical keymap so the dashboard honours
	// the same arrow+vim bindings as every other TUI. Note hjkl are movement
	// here (via Up/Down/Left/Right), so tiles whose hotkey collides with them
	// stay shadowed — matching prior behaviour.
	keys := DefaultKeys()
	switch {
	case key.Matches(msg, keys.Quit), key.Matches(msg, keys.Back):
		m.result = DashboardResult{Quit: true}
		m.done = true
		return m, tea.Quit
	case key.Matches(msg, keys.Filter):
		m.searchOn = true
		m.query = ""
		m.rebuildVisible()
		return m, nil
	case key.Matches(msg, keys.Up):
		m = m.moveCursor(-m.cols())
		return m, nil
	case key.Matches(msg, keys.Down):
		m = m.moveCursor(m.cols())
		return m, nil
	case key.Matches(msg, keys.Left):
		m = m.moveCursor(-1)
		return m, nil
	case key.Matches(msg, keys.Right):
		m = m.moveCursor(1)
		return m, nil
	case key.Matches(msg, keys.Enter):
		return m.launchCursor()
	}

	// Keys outside the shared vocabulary (grid paging, tab cycling, per-tile
	// hotkeys) stay literal.
	k := msg.String()
	switch k {
	case "tab":
		m = m.moveCursor(1)
		return m, nil
	case "shift+tab":
		m = m.moveCursor(-1)
		return m, nil
	case "pgup":
		if m.ready {
			m.vp.ViewUp()
		}
		return m, nil
	case "pgdown":
		if m.ready {
			m.vp.ViewDown()
		}
		return m, nil
	default:
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

// gridHeight is the number of rows available to the scrollable grid view.
// Total vertical budget = height. Non-grid lines: header(2) + blank(1) +
// search(3) + blank(1) + footer(2) = 9. So grid gets whatever is left.
func (m dashboardModel) gridHeight() int {
	h := m.height - 9
	if h < 3 {
		h = 3
	}
	return h
}

// syncViewport rebuilds the grid content, resizes if needed, and scrolls so
// the cursor tile is on-screen.
func (m dashboardModel) syncViewport() dashboardModel {
	if !m.ready {
		return m
	}
	m.vp.Width = m.width
	m.vp.Height = m.gridHeight()
	m.vp.SetContent(m.renderGrid())
	m = m.ensureCursorVisible()
	return m
}

func (m dashboardModel) resizeViewport() dashboardModel {
	if m.width == 0 || m.height == 0 {
		return m
	}
	if !m.ready {
		m.vp = viewport.New(m.width, m.gridHeight())
		m.ready = true
	}
	return m.syncViewport()
}

// ensureCursorVisible nudges the viewport so the tile under the cursor sits
// inside the visible window. Rows are uniform height (tileDisplayHeight) so we
// can compute line ranges arithmetically.
func (m dashboardModel) ensureCursorVisible() dashboardModel {
	if !m.ready || len(m.visible) == 0 {
		return m
	}
	row := m.cursor / m.cols()
	tileH := tileDisplayHeight
	top := row * tileH
	bottom := top + tileH - 1
	off := m.vp.YOffset
	vh := m.vp.Height
	if top < off {
		m.vp.SetYOffset(top)
	} else if bottom >= off+vh {
		m.vp.SetYOffset(bottom - vh + 1)
	}
	return m
}

func (m dashboardModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	th := m.cfg.Theme
	header := Header(th, HeaderSpec{
		Version: m.cfg.Version,
		Section: "Dashboard",
		Meta:    RenderStatusMeta(th, m.cfg.Status),
	}, m.width)

	right := th.Meta.Render("focused: ") + th.Brand.Render(m.focusedLabel())
	footer := Footer(th, m.hints(), right, m.width)

	search := indentBlock(m.renderSearchBar(), 2)

	gridView := ""
	if m.ready {
		gridView = indentBlock(m.vp.View(), 2)
	}

	if len(m.visible) == 0 {
		gridView += "\n  " + th.Crumb.Render("no command matches "+fmt.Sprintf("%q", m.query))
	}

	// Pad the grid view to exactly gridHeight() lines so the footer sits at
	// a predictable row regardless of content size.
	gridView = padToHeight(gridView, m.gridHeight())

	return header + "\n\n" + search + "\n\n" + gridView + "\n" + footer
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

// RenderStatusMeta is the canonical "● healthy · N platforms · M skills"
// string shown in the header Meta slot. Shared by the dashboard and every
// sub-screen so the top-right summary stays consistent as the user navigates.
func RenderStatusMeta(theme *ui.Theme, status DashboardStatus) string {
	if theme == nil {
		theme = ui.DefaultTheme()
	}
	dot := theme.DotOK
	label := "healthy"
	if !status.Healthy {
		dot = theme.DotWarn
		label = "drift"
	}
	sep := theme.Crumb.Render(" · ")
	return dot.Render("●") + " " + theme.Detail.Render(label) +
		sep + theme.Detail.Render(fmt.Sprintf("%d platform%s", status.Platforms, pluralDash(status.Platforms))) +
		sep + theme.Detail.Render(fmt.Sprintf("%d skill%s", status.Skills, pluralDash(status.Skills)))
}

// bodyWidth is the usable width between left/right margins. Every body
// element (search bar, tile grid row) targets this width so they all end at
// the same column.
func (m dashboardModel) bodyWidth() int {
	w := m.width - 4
	if w < 40 {
		w = 40
	}
	return w
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
	var queryRaw string
	var queryStyle lipgloss.Style
	switch {
	case m.query == "" && !m.searchOn:
		queryRaw, queryStyle = "press / to search", th.Crumb
	case m.query == "":
		queryRaw, queryStyle = "type to filter…", th.Crumb
	default:
		queryRaw, queryStyle = m.query, th.Name
	}
	countRaw := fmt.Sprintf("%d / %d", len(m.visible), len(m.cfg.Tiles))
	countStyled := th.Crumb.Render(countRaw)

	prefix := sigil + "  "
	prefixW := lipgloss.Width(prefix)
	countW := lipgloss.Width(countStyled)

	// Pack content to `budget = inner - 2`, giving a 2-cell safety margin
	// inside the outer Width(inner). Some glyphs (notably ❯ U+276F) are
	// reported as 1-cell by runewidth but render 2 cells in several
	// terminals — packing to exactly `inner` then wraps the search bar to
	// a 2nd line, which shoves the header off-screen on alt-screen. The
	// buffer ensures the outer `.Width(inner)` always pads (never wraps).
	budget := inner - 2
	if budget < prefixW+1 {
		budget = prefixW + 1
	}
	var line string
	if budget-prefixW-countW-1 < 1 {
		line = prefix + queryStyle.Render(truncateDisplay(queryRaw, budget-prefixW))
	} else {
		maxQueryW := budget - prefixW - countW - 1
		q := truncateDisplay(queryRaw, maxQueryW)
		pad := budget - prefixW - lipgloss.Width(q) - countW
		if pad < 1 {
			pad = 1
		}
		line = prefix + queryStyle.Render(q) + strings.Repeat(" ", pad) + countStyled
	}
	borderColor := th.Palette.Border
	if m.searchOn {
		borderColor = th.Palette.Magenta
	}
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(inner).
		Render(line)
}

// tileDisplayHeight is the fixed rendered height of every tile, in lines:
// border(2) + header(1) + desc(1) + foot(1) = 5. Staying uniform lets us
// compute row offsets for scroll-to-cursor.
const tileDisplayHeight = 5

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
		sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, parts...))
		if i < len(rows)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// renderTile renders one tile. The returned string is exactly `width`
// display columns wide and tileDisplayHeight lines tall so the grid is a
// perfect lattice.
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

	desc := th.Desc.Render(truncateDisplay(t.Desc, inner))

	footLeft := th.Detail.Render(truncateDisplay(t.Sub, inner))
	footRight := ""
	if t.Status != "" {
		footRight = th.BadgeGhost.Render(t.Status)
	}
	foot := padBetween(footLeft, footRight, inner)

	body := header + "\n" + desc + "\n" + foot
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(width - 2).
		Render(body)
}

// renderEmptyTile returns a blank block sized like a real tile so short final
// rows still align under a full row above.
func (m dashboardModel) renderEmptyTile(width int) string {
	return lipgloss.NewStyle().
		Width(width).
		Height(tileDisplayHeight).
		Render("")
}

// indentBlock prefixes every line of s with n spaces. Using a naked prefix
// ("  " + s) only shifts line one and leaves the remaining rows at col 0 —
// that's the off-center bug that made card tops look pushed to the right.
func indentBlock(s string, n int) string {
	if n <= 0 || s == "" {
		return s
	}
	pad := strings.Repeat(" ", n)
	return pad + strings.ReplaceAll(s, "\n", "\n"+pad)
}

// truncateDisplay clips s to at most width display cells, appending an ellipsis
// so the tile rows stay single-line and the grid stays a uniform lattice.
func truncateDisplay(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width <= 1 {
		return "…"
	}
	runes := []rune(s)
	// Trim rune-by-rune until we fit with room for the ellipsis.
	for len(runes) > 0 && lipgloss.Width(string(runes))+1 > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

func pluralDash(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
