package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// BrowseItem is the contract for a row shown in the browser. Commands
// implement it over their domain type (registry.Skill for search, a manifest
// entry for list) so a single model drives both surfaces.
type BrowseItem interface {
	list.Item
	Title() string
	Description() string
	// Preview returns the right-pane content. width is the viewport's usable
	// width — implementations should wrap to honour it.
	Preview(theme *ui.Theme, width int) string
}

// BrowseAction names the shortcut the user pressed to exit the browser.
type BrowseAction string

const (
	ActionNone      BrowseAction = ""
	ActionInstall   BrowseAction = "install"
	ActionUpdate    BrowseAction = "update"
	ActionUninstall BrowseAction = "uninstall"
	ActionQuit      BrowseAction = "quit"
)

// BrowserConfig parametrises NewBrowser.
type BrowserConfig struct {
	// Command is the breadcrumb label, e.g. "search" or "installed".
	Command string
	// Detail optionally appears after the breadcrumb, e.g. a query.
	Detail string
	// Theme drives every rendered colour.
	Theme *ui.Theme
	// Items populate the left-side list.
	Items []BrowseItem
	// Actions is the set of action shortcuts that should exit the browser.
	// Include ActionInstall in search mode; ActionUpdate + ActionUninstall in
	// installed-list mode.
	Actions []BrowseAction
	// EmptyMsg is shown when Items is empty.
	EmptyMsg string
}

// BrowseResult is what Browser.Selected returns after Run exits.
type BrowseResult struct {
	Action BrowseAction
	Item   BrowseItem
}

// Browser is a bubbletea model: left-side filterable list, right-side preview
// pane, shared header/footer chrome. All rendering honours cfg.Theme so it
// looks identical to every other humblskills TUI surface.
type Browser struct {
	cfg      BrowserConfig
	list     list.Model
	preview  viewport.Model
	width    int
	height   int
	result   BrowseResult
	keys     Keys
	actionOn map[BrowseAction]bool
}

// NewBrowser builds a Browser ready to hand to Run. Callers typically do:
//
//	res, err := tui.Run(tui.NewBrowser(cfg))
//	m := res.(tui.Browser)
//	switch m.Selected().Action { ... }
func NewBrowser(cfg BrowserConfig) Browser {
	items := make([]list.Item, len(cfg.Items))
	for i, it := range cfg.Items {
		items[i] = it
	}

	delegate := newBrowseDelegate(cfg.Theme)
	l := list.New(items, delegate, 0, 0)
	l.Title = cfg.EmptyMsg // reused as header message slot when empty
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.SetShowFilter(true)
	l.SetShowPagination(true)
	l.DisableQuitKeybindings()

	actionOn := map[BrowseAction]bool{}
	for _, a := range cfg.Actions {
		actionOn[a] = true
	}

	vp := viewport.New(0, 0)
	return Browser{
		cfg:      cfg,
		list:     l,
		preview:  vp,
		keys:     DefaultKeys(),
		actionOn: actionOn,
	}
}

// Selected returns the chosen item + action once the browser has exited.
func (b Browser) Selected() BrowseResult { return b.result }

func (b Browser) Init() tea.Cmd { return nil }

func (b Browser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width, b.height = msg.Width, msg.Height
		b.resize()
		b.refreshPreview()
		return b, nil

	case tea.KeyMsg:
		// Never steal keys while the user is typing in the filter input.
		if b.list.FilterState() == list.Filtering {
			break
		}
		switch {
		case key.Matches(msg, b.keys.Quit):
			b.result = BrowseResult{Action: ActionQuit}
			return b, tea.Quit
		case key.Matches(msg, b.keys.Enter):
			if b.actionOn[ActionInstall] {
				return b.exitWith(ActionInstall)
			}
			if b.actionOn[ActionUpdate] {
				return b.exitWith(ActionUpdate)
			}
		}
		switch msg.String() {
		case "i":
			if b.actionOn[ActionInstall] {
				return b.exitWith(ActionInstall)
			}
		case "u":
			if b.actionOn[ActionUpdate] {
				return b.exitWith(ActionUpdate)
			}
		case "x":
			if b.actionOn[ActionUninstall] {
				return b.exitWith(ActionUninstall)
			}
		}
	}

	var cmd tea.Cmd
	prevIdx := b.list.Index()
	b.list, cmd = b.list.Update(msg)
	if b.list.Index() != prevIdx {
		b.refreshPreview()
	}
	return b, cmd
}

func (b Browser) View() string {
	th := b.cfg.Theme
	header := Header(th, b.cfg.Command, b.cfg.Detail, b.width)

	body := b.renderBody()
	footer := Footer(th, b.hints())
	bodyHeight := b.height - lipgloss.Height(header) - lipgloss.Height(footer) - 3
	if bodyHeight < 5 {
		bodyHeight = 5
	}
	return Frame(header, body, footer, bodyHeight)
}

func (b Browser) renderBody() string {
	if len(b.cfg.Items) == 0 {
		return "  " + b.cfg.Theme.Detail.Render(b.cfg.EmptyMsg)
	}
	left := b.list.View()
	right := b.preview.View()
	row := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	return row
}

func (b Browser) hints() []KeyHint {
	hints := []KeyHint{
		{Keys: "↑/↓", Label: "navigate"},
		{Keys: "/", Label: "filter"},
	}
	if b.actionOn[ActionInstall] {
		hints = append(hints, KeyHint{Keys: "enter/i", Label: "install"})
	}
	if b.actionOn[ActionUpdate] {
		hints = append(hints, KeyHint{Keys: "u", Label: "update"})
	}
	if b.actionOn[ActionUninstall] {
		hints = append(hints, KeyHint{Keys: "x", Label: "uninstall"})
	}
	hints = append(hints, KeyHint{Keys: "q", Label: "quit"})
	return hints
}

func (b *Browser) exitWith(a BrowseAction) (tea.Model, tea.Cmd) {
	selected, ok := b.list.SelectedItem().(BrowseItem)
	if !ok {
		return *b, nil
	}
	b.result = BrowseResult{Action: a, Item: selected}
	return *b, tea.Quit
}

func (b *Browser) resize() {
	if b.width == 0 || b.height == 0 {
		return
	}
	// Header (2 lines) + blank + footer (1 line) + blank = 5 rows of chrome.
	bodyHeight := b.height - 6
	if bodyHeight < 5 {
		bodyHeight = 5
	}
	leftW := b.width / 3
	if leftW < 24 {
		leftW = 24
	}
	if leftW > 44 {
		leftW = 44
	}
	rightW := b.width - leftW - 4
	if rightW < 20 {
		rightW = 20
	}
	b.list.SetSize(leftW, bodyHeight)
	b.preview.Width = rightW
	b.preview.Height = bodyHeight
}

func (b *Browser) refreshPreview() {
	it, ok := b.list.SelectedItem().(BrowseItem)
	if !ok {
		b.preview.SetContent("")
		return
	}
	b.preview.SetContent(it.Preview(b.cfg.Theme, b.preview.Width))
	b.preview.GotoTop()
}

// browseDelegate renders each row with our palette. Two lines per row: bold
// title + muted description. On selection the bullet + title pick up brand
// colour so the active row is unambiguous.
type browseDelegate struct {
	theme *ui.Theme
}

func newBrowseDelegate(t *ui.Theme) browseDelegate { return browseDelegate{theme: t} }

func (d browseDelegate) Height() int  { return 2 }
func (d browseDelegate) Spacing() int { return 1 }

func (d browseDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d browseDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it, ok := item.(BrowseItem)
	if !ok {
		return
	}
	selected := index == m.Index()

	title := it.Title()
	desc := it.Description()

	var titleStyle, descStyle lipgloss.Style
	var bullet string
	if selected {
		bullet = d.theme.Bullet.Render("▸ ")
		titleStyle = d.theme.Name
		descStyle = d.theme.Detail
	} else {
		bullet = "  "
		titleStyle = d.theme.Info.Bold(false)
		descStyle = d.theme.Detail
	}

	fmt.Fprintln(w, bullet+titleStyle.Render(title))
	if desc != "" {
		fmt.Fprintln(w, "  "+descStyle.Render(trimOne(desc)))
	} else {
		fmt.Fprintln(w, "")
	}
}

// trimOne clips to one line so the two-line delegate height stays consistent.
func trimOne(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
