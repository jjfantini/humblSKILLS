package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/v2/internal/textutil"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/ui"
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

// HeaderItem marks a non-selectable section-header row (e.g. a "── group ──"
// divider). Header rows are rendered in place but skipped during navigation and
// never returned as a Result. While a filter is active, headers are hidden and
// the list collapses to a flat matched view.
type HeaderItem interface {
	Item
	IsHeader() bool
}

// CollapsibleItem marks a focusable section header (category/role) that can
// hide its descendants when collapsed. Unlike HeaderItem, these rows ARE
// navigable; Enter toggles collapse instead of firing the first Action.
type CollapsibleItem interface {
	Item
	IsCollapsible() bool
	CollapseKey() string
	// RowCollapsed is like Row but includes expand/collapse caret state.
	RowCollapsed(theme *ui.Theme, width int, selected, collapsed bool) string
}

// NestedItem reports ancestor collapse keys. When any key is collapsed in the
// model's map, the item is hidden from the visible list.
type NestedItem interface {
	Item
	ParentCollapseKeys() []string
}

func isHeader(it Item) bool {
	h, ok := it.(HeaderItem)
	return ok && h.IsHeader()
}

func isCollapsible(it Item) bool {
	c, ok := it.(CollapsibleItem)
	return ok && c.IsCollapsible()
}

// isNavSkip reports whether ↑/↓ should skip this row. Collapsible headers are
// focusable; plain HeaderItem dividers are not.
func isNavSkip(it Item) bool {
	if isCollapsible(it) {
		return false
	}
	return isHeader(it)
}

func parentCollapseKeys(it Item) []string {
	if n, ok := it.(NestedItem); ok {
		return n.ParentCollapseKeys()
	}
	return nil
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
	cfg       Config
	items     []Item // visible view (filter + collapse applied)
	cursor    int
	width     int
	height    int
	preview   viewport.Model
	filter    textinput.Model
	filtOn    bool
	helpOn    bool // ? overlay: full-body keybinding cheatsheet
	result    Result
	keys      Keys
	actions   map[string]ActionSpec // keyed by ActionSpec.Key
	done      bool
	collapsed map[string]bool // CollapseKey → collapsed; missing = expanded
	// listOff is the index of the first item row the left pane draws and
	// listRows how many rows fit below its title. The left pane builds a plain
	// string rather than owning a viewport, so without this window renderLeft
	// emitted every item: a list taller than the terminal overflowed the frame,
	// the alt-screen scrolled the header off the top, and no key brought it
	// back. listOff is never assigned from the outside — scrollCursorIntoView
	// derives it from the cursor, so "where the list is scrolled to" has a
	// single owner and can't drift from the highlighted row.
	listOff  int
	listRows int
}

// wheelStep is how many list rows one mouse-wheel notch travels.
const wheelStep = 3

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
		cfg:       cfg,
		preview:   vp,
		filter:    filt,
		keys:      DefaultKeys(),
		actions:   acts,
		collapsed: map[string]bool{},
	}
	m.rebuildVisible()
	m.cursor = m.firstSelectable()
	return m
}

// firstSelectable returns the index of the first navigable item, or 0.
func (m Model) firstSelectable() int {
	for i := range m.items {
		if !isNavSkip(m.items[i]) {
			return i
		}
	}
	return 0
}

// nextSelectable / prevSelectable return the next navigable index in the given
// direction, or -1 if there is none.
func (m Model) nextSelectable(from int) int {
	for i := from + 1; i < len(m.items); i++ {
		if !isNavSkip(m.items[i]) {
			return i
		}
	}
	return -1
}

func (m Model) prevSelectable(from int) int {
	for i := from - 1; i >= 0; i-- {
		if !isNavSkip(m.items[i]) {
			return i
		}
	}
	return -1
}

// moveCursor steps the cursor n navigable rows (negative = up), stopping dead at
// either end, then re-syncs the detail pane and the scroll window. Every cursor
// movement goes through here — ↑/↓ and the mouse wheel — so there is exactly one
// place that has to remember to scroll the new row into view.
func (m *Model) moveCursor(n int) {
	if n == 0 {
		return
	}
	up := n < 0
	if up {
		n = -n
	}
	for i := 0; i < n; i++ {
		next := m.nextSelectable(m.cursor)
		if up {
			next = m.prevSelectable(m.cursor)
		}
		if next < 0 {
			break
		}
		m.cursor = next
	}
	m.scrollCursorIntoView()
	m.refreshPreview()
}

// scrollCursorIntoView slides listOff the minimum distance needed to keep the
// cursor row inside the visible window, so the list scrolls one row at a time at
// the edges instead of jumping the selection to the middle. Call after anything
// that moves the cursor, changes which items are visible, or resizes the pane.
func (m *Model) scrollCursorIntoView() {
	// listRows is 0 until the first WindowSizeMsg, and a list that already fits
	// never scrolls; both mean "draw from the top".
	if m.listRows < 1 || len(m.items) <= m.listRows {
		m.listOff = 0
		return
	}
	if maxOff := len(m.items) - m.listRows; m.listOff > maxOff {
		m.listOff = maxOff
	}
	if m.listOff < 0 {
		m.listOff = 0
	}
	if m.cursor < m.listOff {
		m.listOff = m.cursor
	}
	if m.cursor >= m.listOff+m.listRows {
		m.listOff = m.cursor - m.listRows + 1
	}
}

// listWindow is the [start, end) span of m.items the left pane renders. It
// re-clamps rather than trusting listOff because View runs on a value copy that
// may not have seen a resize yet.
func (m Model) listWindow() (start, end int) {
	if m.listRows < 1 || len(m.items) <= m.listRows {
		return 0, len(m.items)
	}
	start = m.listOff
	if maxOff := len(m.items) - m.listRows; start > maxOff {
		start = maxOff
	}
	if start < 0 {
		start = 0
	}
	end = start + m.listRows
	if end > len(m.items) {
		end = len(m.items)
	}
	return start, end
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
		// The pane under the pointer owns the wheel, so a trackpad flick over
		// the list doesn't also scroll the detail body. Over the left pane a
		// notch moves the cursor by wheelStep rows — same path as holding ↑/↓,
		// which keeps the highlight, the detail pane and the scroll window in
		// lockstep instead of letting the list scroll away from its selection.
		leftW, _ := m.paneWidths()
		if msg.X < leftW {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				m.moveCursor(-wheelStep)
			case tea.MouseButtonWheelDown:
				m.moveCursor(wheelStep)
			}
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
		m.moveCursor(-1)
		return m, nil
	case key.Matches(msg, m.keys.Down):
		m.moveCursor(1)
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
		// Collapsible section headers absorb enter to toggle expand/collapse.
		if len(m.items) > 0 && m.cursor < len(m.items) && isCollapsible(m.items[m.cursor]) {
			c := m.items[m.cursor].(CollapsibleItem)
			key := c.CollapseKey()
			m.collapsed[key] = !m.collapsed[key]
			focusKey := key
			m.rebuildVisible()
			m.cursor = m.indexOfCollapseKey(focusKey)
			// Collapsing removes rows above the cursor and expanding adds rows
			// below it; either way the window has to re-centre on the header the
			// user just toggled.
			m.scrollCursorIntoView()
			m.refreshPreview()
			return m, nil
		}
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
	// Never surface a header or collapsible section row as a selection.
	if isHeader(it) || isCollapsible(it) {
		it = nil
	}
	m.result = Result{Action: action, Item: it}
	m.done = true
	return m, tea.Quit
}

// indexOfCollapseKey finds a collapsible header by CollapseKey in the visible
// list, or falls back to firstSelectable.
func (m Model) indexOfCollapseKey(key string) int {
	for i, it := range m.items {
		if c, ok := it.(CollapsibleItem); ok && c.IsCollapsible() && c.CollapseKey() == key {
			return i
		}
	}
	return m.firstSelectable()
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
	// The left pane spends its budget the same way: title row, blank row, then
	// item rows. Same subtraction keeps the two panes row-synced.
	m.listRows = bodyH - 2
	if m.listRows < 1 {
		m.listRows = 1
	}
	m.scrollCursorIntoView()
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
	m.rebuildVisible()
	// Keep the cursor in range and off non-navigable rows.
	if len(m.items) == 0 {
		m.cursor = 0
	} else if m.cursor >= len(m.items) || isNavSkip(m.items[m.cursor]) {
		m.cursor = m.firstSelectable()
	}
	// Typing into the filter shrinks the list under the window, so the offset
	// has to come back down with it or the pane renders past the last match.
	m.scrollCursorIntoView()
}

// rebuildVisible rebuilds m.items from cfg.Items applying the active filter
// and collapse state. While filtering, all headers (including collapsible
// section headers) are hidden so the list is a flat match view.
func (m *Model) rebuildVisible() {
	q := strings.ToLower(strings.TrimSpace(m.filter.Value()))
	if q != "" {
		out := make([]Item, 0, len(m.cfg.Items))
		for _, it := range m.cfg.Items {
			if isHeader(it) || isCollapsible(it) {
				continue
			}
			if strings.Contains(strings.ToLower(it.FilterValue()), q) {
				out = append(out, it)
			}
		}
		m.items = out
		return
	}
	if m.collapsed == nil {
		m.collapsed = map[string]bool{}
	}
	out := make([]Item, 0, len(m.cfg.Items))
	for _, it := range m.cfg.Items {
		if hiddenByCollapse(it, m.collapsed) {
			continue
		}
		out = append(out, it)
	}
	m.items = out
}

func hiddenByCollapse(it Item, collapsed map[string]bool) bool {
	for _, k := range parentCollapseKeys(it) {
		if collapsed[k] {
			return true
		}
	}
	return false
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

	start, end := m.listWindow()

	var title string
	if m.filtOn {
		m.filter.Width = width - 2
		title = "  " + m.filter.View()
	} else {
		title = "  " + th.SectionTitle.Render(spacedUpper(m.cfg.LeftTitle))
		// Right-anchor the overflow affordance on the title row, mirroring the
		// detail pane's indicator. Skipped while filtering — the title row is
		// the text input then, and the match count in the footer already says
		// how much the list holds.
		if ind := m.listScrollIndicator(th, start, end); ind != "" {
			title = padBetween(title, ind, width-1)
		}
	}

	if len(m.items) == 0 {
		empty := "  " + th.Detail.Render(textutil.FirstNonEmpty(m.cfg.EmptyMsg, "— no items —"))
		return pad(title) + "\n\n" + pad(empty)
	}

	var sb strings.Builder
	sb.WriteString(pad(title))
	sb.WriteString("\n\n")
	for i := start; i < end; i++ {
		it := m.items[i]
		selected := i == m.cursor
		var row string
		if c, ok := it.(CollapsibleItem); ok && c.IsCollapsible() {
			row = c.RowCollapsed(th, width-2, selected, m.collapsed[c.CollapseKey()])
		} else {
			row = it.Row(th, width-2, selected)
		}
		// Fit the row body (minus the leading bar/gutter) to exactly width-2.
		// Truncating matters as much as padding: Item.Row is free to return a
		// row wider than the pane (skillItem does, for a long name plus its
		// version), and lipgloss.JoinHorizontal sizes the left block to its
		// widest line. Pad-only meant one long row shoved the divider — and the
		// whole detail pane — right by however far it overhung, so with the
		// scroll window in play the divider jittered as rows entered and left
		// the view. Clamping here fixes it for every Item type at once.
		row = fitToWidth(row, width-2)
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
		if i < end-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// listScrollIndicator is the left pane's counterpart to scrollIndicator: hollow
// arrows at the ends, filled when rows are hidden that way, plus how far through
// the list the window sits. Empty string when everything already fits.
func (m Model) listScrollIndicator(th *ui.Theme, start, end int) string {
	if start <= 0 && end >= len(m.items) {
		return ""
	}
	up, down := "△", "▽"
	if start > 0 {
		up = "▲"
	}
	if end < len(m.items) {
		down = "▼"
	}
	// Percent of the way scrolled, not percent of rows seen: hidden is the
	// number of scroll positions available, so the first row reads 0% and the
	// last reads 100% — matching viewport.ScrollPercent on the detail side.
	pct := 100
	if hidden := len(m.items) - m.listRows; hidden > 0 {
		pct = int(float64(start)/float64(hidden)*100 + 0.5)
	}
	return th.Meta.Render(fmt.Sprintf("%s%s %d%%", up, down, pct))
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

	onSection := len(m.items) > 0 && m.cursor < len(m.items) && isCollapsible(m.items[m.cursor])
	if onSection {
		hints = append(hints, KeyHint{Keys: "enter", Label: "toggle"})
	}

	// Deduplicate enter when it's absorbed by the first action.
	seen := map[string]bool{}
	if onSection {
		seen["enter"] = true // enter already claimed by section toggle
	}
	for _, a := range m.cfg.Actions {
		label := a.Label
		keyStr := a.Key
		if !onSection && (keyStr == "enter" || (len(m.cfg.Actions) > 0 && !seen["enter"] && a.Key == m.cfg.Actions[0].Key)) {
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
		{"wheel", "scroll the list (pointer over the left pane)"},
		{"enter", "toggle section (on category/role headers)"},
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

	// Column split: movement concerns on the left, what-you-can-do on the right.
	// Stacked in one column the sheet runs past 20 rows once a caller supplies a
	// few Actions, which overflows a stock 24-row terminal — and an overflowing
	// body now loses its tail to Frame's clamp instead of shoving the header off
	// screen. Two columns roughly halve the height and keep it all readable.
	leftCol := []helpGroup{nav, scroll}
	rightCol := make([]helpGroup, 0, 2)

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
		rightCol = append(rightCol, helpGroup{title: "ACTIONS", rows: rows})
	}

	backKey := textutil.FirstNonEmpty(m.cfg.BackKey, "q")
	backLabel := textutil.FirstNonEmpty(m.cfg.BackLabel, "quit")
	// esc always goes back, so a caller that already bound BackKey to esc would
	// otherwise render "esc / esc".
	backKeys := backKey + " / esc"
	if backKey == "esc" {
		backKeys = "esc"
	}
	rightCol = append(rightCol, helpGroup{title: "GENERAL", rows: []helpRow{
		{"?", "toggle this help"},
		{backKeys, backLabel},
		{"ctrl+c", "quit"},
	}})

	// Align the key column to the widest key across every group so the labels
	// form a clean second column.
	keyW := 0
	for _, g := range append(append([]helpGroup{}, leftCol...), rightCol...) {
		for _, r := range g.rows {
			if w := lipgloss.Width(r.keys); w > keyW {
				keyW = w
			}
		}
	}

	renderCol := func(groups []helpGroup) string {
		var sb strings.Builder
		for i, g := range groups {
			if i > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(th.Meta.Render(g.title) + "\n")
			for _, r := range g.rows {
				keyCol := r.keys + strings.Repeat(" ", keyW-lipgloss.Width(r.keys))
				sb.WriteString("  " + th.Brand.Render(keyCol) + "  " + th.Detail.Render(r.label) + "\n")
			}
		}
		return strings.TrimRight(sb.String(), "\n")
	}

	// JoinHorizontal pads each block to its own widest line, so the gutter style
	// is all that separates the columns.
	gutter := lipgloss.NewStyle().PaddingRight(4)
	cols := lipgloss.JoinHorizontal(lipgloss.Top,
		gutter.Render(renderCol(leftCol)),
		renderCol(rightCol),
	)
	return "  " + th.SectionTitle.Render(spacedUpper("Keybindings")) + "\n\n" + indentBlock(cols, 2)
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
