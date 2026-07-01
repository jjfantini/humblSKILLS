package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/internal/textutil"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// Item is the contract every command's list row satisfies. One Item per
// rendered row in the left pane.
type Item interface {
	// Key is the stable identifier (name, slug). Shown in the header's right
	// meta as "focused: <key>".
	Key() string
	// Row returns the left-pane body for this item. The model overlays a
	// leading magenta ▌ bar for the cursor row; the Item itself decides how
	// to style the name / dot / badge based on `selected`. width is the
	// usable columns inside the left pane (not counting the bar).
	Row(theme *ui.Theme, width int, selected bool) string
	// Detail returns the right-pane body for this item. width is the usable
	// columns inside the right pane.
	Detail(theme *ui.Theme, width int) string
	// FilterValue is the haystack string the built-in filter matches against.
	FilterValue() string
}

// SizedItem is an optional Item extension that reports the item's *natural*
// left-pane width (in display cells, NOT including the 2-cell gutter the
// model prepends). Items that implement it let the pane snug the divider to
// the content; items that don't fall back to a safe default.
//
// Why an interface instead of measuring Row output: Row() renders with ANSI
// styling (badges with Padding/Background leave trailing `\x1b[0m` rather
// than literal spaces), so strings.TrimRight can't recover the natural width
// from a padded render. Letting the item compute its own width directly
// sidesteps that entirely.
type SizedItem interface {
	Item
	NaturalWidth(theme *ui.Theme) int
}

// ActionSpec binds a key to a caller-named action. Pressing Key while an item
// is highlighted exits the model with Result{Action, Item}.
type ActionSpec struct {
	Key     string
	Label   string
	Action  string
	Enabled func(it Item) bool
}

// Config parametrises NewListDetail.
type Config struct {
	Theme      *ui.Theme
	Section    string // crumb after "humblskills vX.Y.Z" ("Adapters")
	Version    string // e.g. "v0.4.2"
	Meta       func(items []Item, cursor int) string
	Items      []Item
	LeftTitle  string // "ADAPTERS" / "SKILLS" / "INSTALLED"
	RightTitle string // "DETAIL"
	Actions    []ActionSpec
	EmptyMsg   string
	// LeftWidth overrides the computed left-pane width (in cells, including
	// the leading gutter). When 0 the model sizes it to the widest rendered
	// row, capped at width/3. Use this to force a tighter (or wider) column.
	LeftWidth int
	// FocusedLabel overrides the default "focused: <key>" right-anchored footer
	// context. Return "" for no context.
	FocusedLabel func(items []Item, cursor int) string
	// BackLabel overrides the quit-hint label. Default is "quit". Set to
	// "back" (with BackKey = "esc") when the model is launched from a parent
	// navigator so ESC feels like "go back" instead of "quit".
	BackLabel string
	// BackKey overrides the quit-hint key. Default is "q".
	BackKey string
}

// Result is what the caller inspects after the model returns.
type Result struct {
	Action string // "" if user quit
	Item   Item   // nil if Items was empty or user quit
	Quit   bool
}

// Model is the shared two-pane bubbletea model.
type Model struct {
	cfg     Config
	items   []Item // filtered view (equal to cfg.Items when filter is empty)
	cursor  int
	width   int
	height  int
	preview viewport.Model
	filter  textinput.Model
	filtOn  bool
	helpOn  bool // ? overlay: full-body keybinding cheatsheet
	result  Result
	keys    Keys
	actions map[string]ActionSpec // keyed by ActionSpec.Key
	done    bool
}

// NewListDetail builds a Model ready for Run.
func NewListDetail(cfg Config) Model {
	if cfg.Theme == nil {
		cfg.Theme = ui.DefaultTheme()
	}
	if cfg.LeftTitle == "" {
		cfg.LeftTitle = "ITEMS"
	}
	if cfg.RightTitle == "" {
		cfg.RightTitle = "DETAIL"
	}

	filt := textinput.New()
	filt.Prompt = "/ "
	filt.Placeholder = "filter…"
	filt.CharLimit = 64

	vp := viewport.New(0, 0)

	acts := map[string]ActionSpec{}
	for _, a := range cfg.Actions {
		acts[a.Key] = a
	}

	m := Model{
		cfg:     cfg,
		items:   append([]Item(nil), cfg.Items...),
		preview: vp,
		filter:  filt,
		keys:    DefaultKeys(),
		actions: acts,
	}
	return m
}

// Selected returns the terminal state after the model exits.
func (m Model) Selected() Result { return m.result }

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resize()
		m.refreshPreview()
		return m, nil

	case tea.MouseMsg:
		// Forward wheel events to the detail viewport only when the cursor
		// is over the right pane. The left pane's list has its own ↑/↓
		// handling so a trackpad scroll over the list shouldn't also scroll
		// the detail body.
		leftW, _ := m.paneWidths()
		if msg.X < leftW {
			return m, nil
		}
		var cmd tea.Cmd
		m.preview, cmd = m.preview.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if m.helpOn {
			return m.updateHelp(msg)
		}
		if m.filtOn {
			return m.updateFilter(msg)
		}
		return m.updateNav(msg)
	}
	return m, nil
}

// updateHelp handles keys while the ? overlay is open. ctrl+c still quits the
// program; every other key just dismisses the overlay so it never traps the
// user.
func (m Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		m.result = Result{Quit: true}
		m.done = true
		return m, tea.Quit
	}
	m.helpOn = false
	return m, nil
}

func (m Model) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.filtOn = false
		m.filter.Blur()
		m.filter.SetValue("")
		m.applyFilter()
		m.refreshPreview()
		return m, nil
	case "enter":
		m.filtOn = false
		m.filter.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	m.applyFilter()
	m.refreshPreview()
	return m, cmd
}

func (m Model) updateNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit), key.Matches(msg, m.keys.Back):
		m.result = Result{Quit: true}
		m.done = true
		return m, tea.Quit
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
			m.refreshPreview()
		}
		return m, nil
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.items)-1 {
			m.cursor++
			m.refreshPreview()
		}
		return m, nil
	case key.Matches(msg, m.keys.Filter):
		if m.cfg.Items != nil {
			m.filtOn = true
			m.filter.Focus()
		}
		return m, nil
	case key.Matches(msg, m.keys.Help):
		m.helpOn = true
		return m, nil
	}

	// Detail-pane scroll. List navigation owns up/down, so scrolling uses a
	// distinct set: pgup/pgdown for full pages, ctrl+u/ctrl+d for half pages
	// (vim), shift+up/shift+down for single lines. Home/end jump to edges.
	switch msg.String() {
	case "pgup":
		m.preview.ViewUp()
		return m, nil
	case "pgdown":
		m.preview.ViewDown()
		return m, nil
	case "ctrl+u":
		m.preview.HalfViewUp()
		return m, nil
	case "ctrl+d":
		m.preview.HalfViewDown()
		return m, nil
	case "shift+up":
		m.preview.LineUp(1)
		return m, nil
	case "shift+down":
		m.preview.LineDown(1)
		return m, nil
	case "home":
		m.preview.GotoTop()
		return m, nil
	case "end":
		m.preview.GotoBottom()
		return m, nil
	}

	k := msg.String()
	if k == "enter" {
		// First enabled action absorbs enter so users don't need to learn a
		// verb-key for the common case (install, update, apply…).
		for _, a := range m.cfg.Actions {
			if a.Enabled != nil && len(m.items) > 0 && !a.Enabled(m.items[m.cursor]) {
				continue
			}
			return m.exitWith(a.Action)
		}
	}
	if a, ok := m.actions[k]; ok {
		if a.Enabled != nil && len(m.items) > 0 && !a.Enabled(m.items[m.cursor]) {
			return m, nil
		}
		return m.exitWith(a.Action)
	}
	return m, nil
}

func (m Model) exitWith(action string) (tea.Model, tea.Cmd) {
	var it Item
	if len(m.items) > 0 && m.cursor < len(m.items) {
		it = m.items[m.cursor]
	}
	m.result = Result{Action: action, Item: it}
	m.done = true
	return m, tea.Quit
}

func (m *Model) resize() {
	if m.width == 0 || m.height == 0 {
		return
	}
	// Chrome: header (2 lines) + blank + blank + footer (2 lines) = ~6.
	bodyH := m.height - 6
	if bodyH < 5 {
		bodyH = 5
	}
	_, rightW := m.paneWidths()
	// Right pane reserves its first 2 cols for `│ ` — the body divider and
	// a 1-cell gutter before the actual detail content.
	m.preview.Width = rightW - 2
	m.preview.Height = bodyH - 2 // title row + blank row under the title
}

func (m Model) paneWidths() (left, right int) {
	left = m.cfg.LeftWidth
	if left == 0 {
		left = m.measureLeftWidth()
	}
	// Never let the left pane eat more than a third of the screen.
	if cap := m.width / 3; cap > 0 && left > cap {
		left = cap
	}
	if left < 22 {
		left = 22
	}
	// No separator column between panes: the right pane owns its own `│`
	// prefix, so col `leftW` is *the* divider column for every body row
	// AND the column where `DETAIL` starts on the title row. That's the
	// alignment the user asked for: one continuous vertical line from the
	// D of DETAIL down through every `│` below it.
	right = m.width - left
	if right < 20 {
		right = 20
	}
	return left, right
}

// measureLeftWidth computes the natural left-pane width in display cells:
// the max of (section-title width, widest item's NaturalWidth), plus a
// 2-cell gutter, clamped to [minW, maxW].
//
// Items that don't implement SizedItem contribute the fallback width — we
// can't infer their natural width from Row() because the rendered string
// includes ANSI reset sequences that defeat trailing-space trimming. Every
// Item type in this codebase implements SizedItem; the fallback only matters
// for third-party embedders.
func (m Model) measureLeftWidth() int {
	th := m.cfg.Theme
	const (
		minW     = 22
		maxW     = 40
		fallback = 30
		// 2 leading + 1 trailing: "  row " so the widest row ends one cell
		// before `│` and the divider column gets a sliver of breathing room.
		gutter = 3
	)
	widest := minW
	if w := lipgloss.Width(th.SectionTitle.Render(spacedUpper(m.cfg.LeftTitle))) + gutter; w > widest {
		widest = w
	}
	for _, it := range m.cfg.Items {
		var natural int
		if si, ok := it.(SizedItem); ok {
			natural = si.NaturalWidth(th)
		} else {
			natural = fallback
		}
		if w := natural + gutter; w > widest {
			widest = w
		}
	}
	if widest > maxW {
		widest = maxW
	}
	return widest
}

func (m *Model) applyFilter() {
	q := strings.ToLower(strings.TrimSpace(m.filter.Value()))
	if q == "" {
		m.items = append([]Item(nil), m.cfg.Items...)
	} else {
		out := make([]Item, 0, len(m.cfg.Items))
		for _, it := range m.cfg.Items {
			if strings.Contains(strings.ToLower(it.FilterValue()), q) {
				out = append(out, it)
			}
		}
		m.items = out
	}
	if m.cursor >= len(m.items) {
		m.cursor = 0
		if len(m.items) > 0 {
			m.cursor = len(m.items) - 1
			if m.cursor < 0 {
				m.cursor = 0
			}
		}
	}
}

func (m *Model) refreshPreview() {
	if len(m.items) == 0 {
		m.preview.SetContent("")
		return
	}
	it := m.items[m.cursor]
	m.preview.SetContent(it.Detail(m.cfg.Theme, m.preview.Width))
	m.preview.GotoTop()
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	th := m.cfg.Theme
	leftW, rightW := m.paneWidths()

	metaRight := ""
	if m.cfg.Meta != nil {
		metaRight = m.cfg.Meta(m.items, m.cursor)
	}
	header := Header(th, HeaderSpec{
		Version: m.cfg.Version,
		Section: m.cfg.Section,
		Meta:    metaRight,
	}, m.width)

	var body string
	var footer string
	switch {
	case m.helpOn:
		body = m.renderHelp()
		footer = Footer(th, []KeyHint{{Keys: "any key", Label: "close help"}}, "", m.width)
	case m.filtOn:
		// While typing a filter the nav/action keys are inert, so surface the
		// two keys that actually do something (esc clears, enter applies) plus
		// a live match count instead of the stale nav hints.
		body = m.renderBody(leftW, rightW)
		footer = Footer(th, m.filterHints(), m.filterContext(), m.width)
	default:
		body = m.renderBody(leftW, rightW)
		focused := ""
		if m.cfg.FocusedLabel != nil {
			focused = m.cfg.FocusedLabel(m.items, m.cursor)
		} else if len(m.items) > 0 {
			// "focused:" stays muted; the value (item key) renders in the
			// magenta brand colour to match the design.
			focused = th.Meta.Render("focused: ") + th.Brand.Render(m.items[m.cursor].Key())
		}
		footer = Footer(th, m.hints(), focused, m.width)
	}

	bodyH := m.height - lipgloss.Height(header) - lipgloss.Height(footer) - 1
	if bodyH < 5 {
		bodyH = 5
	}
	return Frame(header, body, footer, bodyH)
}

func (m Model) renderBody(leftW, rightW int) string {
	left := m.renderLeft(leftW)
	right := m.renderRight(rightW)
	// No separator block: the right pane renders its own `│` prefix at col
	// 0 on every body row, and `DETAIL` occupies col 0 on the title row.
	// Joining directly puts both at the same absolute column `leftW`.
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m Model) renderLeft(width int) string {
	th := m.cfg.Theme

	// Every line in the left block is padded to exactly `width` cells so
	// lipgloss.JoinHorizontal lines the divider up at a fixed x regardless of
	// which row has the longest content.
	pad := func(s string) string {
		w := lipgloss.Width(s)
		if w >= width {
			return s
		}
		return s + strings.Repeat(" ", width-w)
	}

	var title string
	if m.filtOn {
		m.filter.Width = width - 2
		title = "  " + m.filter.View()
	} else {
		title = "  " + th.SectionTitle.Render(spacedUpper(m.cfg.LeftTitle))
	}

	if len(m.items) == 0 {
		empty := "  " + th.Detail.Render(textutil.FirstNonEmpty(m.cfg.EmptyMsg, "— no items —"))
		return pad(title) + "\n\n" + pad(empty)
	}

	var sb strings.Builder
	sb.WriteString(pad(title))
	sb.WriteString("\n\n")
	for i, it := range m.items {
		selected := i == m.cursor
		row := it.Row(th, width-2, selected)
		// Pad the row body (minus the leading bar/gutter) to width-2.
		rowW := lipgloss.Width(row)
		if rowW < width-2 {
			row += strings.Repeat(" ", width-2-rowW)
		}
		var line string
		if selected {
			// Transparent highlight: just a magenta ▌ bar + magenta-bold
			// name (styled by the Item itself). No background fill so the
			// row stays legible on both dark and light terminal themes.
			bar := th.Bullet.Render("▌")
			line = bar + " " + row
		} else {
			line = "  " + row
		}
		sb.WriteString(line)
		if i < len(m.items)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (m Model) renderRight(width int) string {
	th := m.cfg.Theme
	bar := th.Divider.Render("│")
	title := th.SectionTitle.Render(spacedUpper(m.cfg.RightTitle))

	// Title row has NO `│` prefix — `DETAIL` itself sits in col 0 of the
	// right pane (absolute col `leftW`). Every row below gets a `│` in that
	// same col, so the eye reads `D` at the top as the cap of one unbroken
	// vertical line running down through the body. This is literally the
	// alignment requested: "the split line from the left and right pane
	// lines up with the line that splits SKILLS and DETAIL".
	if len(m.items) == 0 {
		return title
	}

	// When the detail body overflows the viewport, right-anchor a compact
	// scroll indicator on the title row (▲/▽ arrows + percent) so users know
	// there's more to see and which way to scroll.
	if ind := m.scrollIndicator(th); ind != "" {
		title = padBetween(title, ind, width)
	}

	preview := m.preview.View()
	lines := strings.Split(preview, "\n")
	out := make([]string, 0, len(lines)+2)
	out = append(out, title)
	// Blank row between title and preview — mirrors the blank row under
	// `SKILLS` in the left pane so the two panes stay row-synced.
	out = append(out, bar)
	for _, ln := range lines {
		out = append(out, bar+" "+ln)
	}
	return strings.Join(out, "\n")
}

// scrollIndicator returns a compact "▲▼ NN%" widget when the detail viewport
// has more content than fits, or "" when everything is visible. A filled arrow
// means there's content in that direction; a hollow one means you're at that
// edge.
func (m Model) scrollIndicator(th *ui.Theme) string {
	if m.preview.TotalLineCount() <= m.preview.Height {
		return ""
	}
	up, down := "△", "▽"
	if !m.preview.AtTop() {
		up = "▲"
	}
	if !m.preview.AtBottom() {
		down = "▼"
	}
	pct := int(m.preview.ScrollPercent()*100 + 0.5)
	return th.Meta.Render(fmt.Sprintf("%s%s %d%%", up, down, pct))
}

// filterHints are the footer hints shown while the filter input is focused.
// Only esc/enter do anything in this mode, so advertise exactly those.
func (m Model) filterHints() []KeyHint {
	return []KeyHint{
		{Keys: "esc", Label: "clear filter"},
		{Keys: "enter", Label: "apply"},
	}
}

// filterContext is the right-anchored footer text during filtering: a live
// count of how many items currently match.
func (m Model) filterContext() string {
	th := m.cfg.Theme
	n := len(m.items)
	noun := "matches"
	if n == 1 {
		noun = "match"
	}
	return th.Meta.Render(fmt.Sprintf("%d %s", n, noun))
}

func (m Model) hints() []KeyHint {
	hints := []KeyHint{{Keys: "↑↓", Label: "select"}}
	if m.cfg.Items != nil {
		hints = append(hints, KeyHint{Keys: "/", Label: "filter"})
	}
	hints = append(hints, KeyHint{Keys: "⇞⇟", Label: "scroll"})
	// Deduplicate enter when it's absorbed by the first action.
	seen := map[string]bool{}
	for _, a := range m.cfg.Actions {
		label := a.Label
		keyStr := a.Key
		if keyStr == "enter" || (len(m.cfg.Actions) > 0 && !seen["enter"] && a.Key == m.cfg.Actions[0].Key) {
			// First action also triggers on enter.
			keyStr = a.Key + "/enter"
			if a.Key == "enter" {
				keyStr = "enter"
			}
			seen["enter"] = true
		}
		hints = append(hints, KeyHint{Keys: keyStr, Label: label})
	}
	backKey := m.cfg.BackKey
	if backKey == "" {
		backKey = "q"
	}
	backLabel := m.cfg.BackLabel
	if backLabel == "" {
		backLabel = "quit"
	}
	hints = append(hints, KeyHint{Keys: "?", Label: "help"})
	hints = append(hints, KeyHint{Keys: backKey, Label: backLabel})
	return hints
}

// renderHelp draws the ? overlay body: a keybinding cheatsheet grouped by
// concern. It reflects the model's actual configuration (filter only when the
// list is filterable, the caller's Actions, the caller's back key/label) so the
// sheet never advertises a key that does nothing.
func (m Model) renderHelp() string {
	th := m.cfg.Theme

	type helpRow struct{ keys, label string }
	type helpGroup struct {
		title string
		rows  []helpRow
	}

	nav := helpGroup{title: "NAVIGATE", rows: []helpRow{
		{"↑ / k", "move up"},
		{"↓ / j", "move down"},
	}}
	if m.cfg.Items != nil {
		nav.rows = append(nav.rows, helpRow{"/", "filter list"})
	}

	scroll := helpGroup{title: "SCROLL DETAIL", rows: []helpRow{
		{"⇞ / ⇟", "page up / down"},
		{"ctrl+u / ctrl+d", "half page up / down"},
		{"shift+↑ / shift+↓", "line up / down"},
		{"home / end", "jump to top / bottom"},
	}}

	groups := []helpGroup{nav, scroll}

	if len(m.cfg.Actions) > 0 {
		rows := make([]helpRow, 0, len(m.cfg.Actions))
		for i, a := range m.cfg.Actions {
			keys := a.Key
			if i == 0 {
				// The first action is also bound to enter (see updateNav).
				keys = a.Key + " / enter"
			}
			rows = append(rows, helpRow{keys, a.Label})
		}
		groups = append(groups, helpGroup{title: "ACTIONS", rows: rows})
	}

	backKey := textutil.FirstNonEmpty(m.cfg.BackKey, "q")
	backLabel := textutil.FirstNonEmpty(m.cfg.BackLabel, "quit")
	groups = append(groups, helpGroup{title: "GENERAL", rows: []helpRow{
		{"?", "toggle this help"},
		{backKey + " / esc", backLabel},
		{"ctrl+c", "quit"},
	}})

	// Align the key column to the widest key across every group so the labels
	// form a clean second column.
	keyW := 0
	for _, g := range groups {
		for _, r := range g.rows {
			if w := lipgloss.Width(r.keys); w > keyW {
				keyW = w
			}
		}
	}

	var sb strings.Builder
	sb.WriteString("  " + th.SectionTitle.Render(spacedUpper("Keybindings")) + "\n")
	for _, g := range groups {
		sb.WriteString("\n  " + th.Meta.Render(g.title) + "\n")
		for _, r := range g.rows {
			keyCol := r.keys + strings.Repeat(" ", keyW-lipgloss.Width(r.keys))
			sb.WriteString("    " + th.Brand.Render(keyCol) + "  " + th.Detail.Render(r.label) + "\n")
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

// RunListDetail runs the model on an alt-screen and returns the user's choice.
func RunListDetail(cfg Config) (Result, error) {
	m, err := Run(NewListDetail(cfg))
	if err != nil {
		return Result{}, err
	}
	ldm, ok := m.(Model)
	if !ok {
		return Result{}, nil
	}
	return ldm.Selected(), nil
}

// --- small helpers -----------------------------------------------------------

// spacedUpper converts "adapters" → "A D A P T E R S" to match the design's
// tracking-wide section titles (CSS `letter-spacing: 0.18em`). On narrow panes
// falls back to plain uppercase.
func spacedUpper(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	up := strings.ToUpper(s)
	return up
}

func padLeft(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}
