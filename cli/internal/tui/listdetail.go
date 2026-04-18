package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	// FocusedLabel overrides the default "focused: <key>" right-anchored footer
	// context. Return "" for no context.
	FocusedLabel func(items []Item, cursor int) string
}

// Result is what the caller inspects after the model returns.
type Result struct {
	Action string // "" if user quit
	Item   Item   // nil if Items was empty or user quit
	Quit   bool
}

// Model is the shared two-pane bubbletea model.
type Model struct {
	cfg      Config
	items    []Item // filtered view (equal to cfg.Items when filter is empty)
	cursor   int
	width    int
	height   int
	preview  viewport.Model
	filter   textinput.Model
	filtOn   bool
	result   Result
	keys     Keys
	actions  map[string]ActionSpec // keyed by ActionSpec.Key
	done     bool
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

	case tea.KeyMsg:
		if m.filtOn {
			return m.updateFilter(msg)
		}
		return m.updateNav(msg)
	}
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
	case key.Matches(msg, m.keys.Quit):
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
	m.preview.Width = rightW
	m.preview.Height = bodyH - 1 // leave a line for the right-pane section title
}

func (m Model) paneWidths() (left, right int) {
	left = 32
	if l := m.width / 3; l < left {
		left = l
	}
	if left < 22 {
		left = 22
	}
	// Leave: 2-col gutter, left pane, 3-col divider ("  │ " => 1 divider + pads), right pane.
	right = m.width - left - 5
	if right < 20 {
		right = 20
	}
	return left, right
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

	body := m.renderBody(leftW, rightW)

	focused := ""
	if m.cfg.FocusedLabel != nil {
		focused = m.cfg.FocusedLabel(m.items, m.cursor)
	} else if len(m.items) > 0 {
		// "focused:" stays muted; the value (item key) renders in the
		// magenta brand colour to match the design.
		focused = th.Meta.Render("focused: ") + th.Brand.Render(m.items[m.cursor].Key())
	}
	footer := Footer(th, m.hints(), focused, m.width)

	bodyH := m.height - lipgloss.Height(header) - lipgloss.Height(footer) - 1
	if bodyH < 5 {
		bodyH = 5
	}
	return Frame(header, body, footer, bodyH)
}

func (m Model) renderBody(leftW, rightW int) string {
	left := m.renderLeft(leftW)
	divider := m.renderDivider()
	right := m.renderRight(rightW)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, " "+divider+" ", right)
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
		empty := "  " + th.Detail.Render(firstNonEmpty(m.cfg.EmptyMsg, "— no items —"))
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
			bar := th.Bullet.Render("▌")
			// Wrap the content + padding in RowBg so the highlight fills
			// every cell from just after the bar to the divider. The bar
			// itself sits outside the highlight band, matching the design.
			line = bar + th.RowBg.Render(" "+row)
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

func (m Model) renderDivider() string {
	th := m.cfg.Theme
	h := m.preview.Height + 2
	if h < 3 {
		h = 3
	}
	line := strings.Repeat(th.Divider.Render("│")+"\n", h)
	return strings.TrimRight(line, "\n")
}

func (m Model) renderRight(width int) string {
	th := m.cfg.Theme
	title := th.SectionTitle.Render(spacedUpper(m.cfg.RightTitle))
	if len(m.items) == 0 {
		return title
	}
	return title + "\n" + m.preview.View()
}

func (m Model) hints() []KeyHint {
	hints := []KeyHint{{Keys: "↑↓", Label: "select"}}
	if m.cfg.Items != nil {
		hints = append(hints, KeyHint{Keys: "/", Label: "filter"})
	}
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
	hints = append(hints, KeyHint{Keys: "q", Label: "quit"})
	return hints
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

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if v != "" {
			return v
		}
	}
	return ""
}
