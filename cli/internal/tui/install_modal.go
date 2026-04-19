package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// RunInstallPlatformModal asks the user which detected platforms to install
// `skill` into, and at which scope. Default selections come from the profile.
// It is a hand-rolled bubbletea model so ESC navigates back cleanly and the
// chrome matches the rest of the TUI (no huh flicker, no alt-screen re-entry).
//
// Returns Confirmed=false if the user cancelled, and EditProfile=true if they
// chose the "edit profile" action.
func RunInstallPlatformModal(
	theme *ui.Theme,
	skill string,
	adapterList []adapters.Adapter,
	detected map[string]bool,
	p *profile.Profile,
) (InstallModalResult, error) {
	if p == nil {
		p = &profile.Profile{}
	}

	selected := map[string]bool{}
	if len(p.DefaultPlatforms) > 0 {
		for _, name := range p.DefaultPlatforms {
			if detected[name] {
				selected[name] = true
			}
		}
	} else {
		for _, a := range adapterList {
			if detected[a.Name] {
				selected[a.Name] = true
			}
		}
	}

	scopes := []scopeOpt{
		{label: "adapter default", value: ""},
		{label: "user", value: "user"},
		{label: "project", value: "project"},
	}
	scopeIdx := 0
	for i, s := range scopes {
		if s.value == p.DefaultScope {
			scopeIdx = i
			break
		}
	}

	actions := []actionOpt{
		{label: "install", value: "install"},
		{label: "edit profile defaults", value: "profile"},
		{label: "cancel", value: "cancel"},
	}

	m := installModalModel{
		theme:       theme,
		skill:       skill,
		adapters:    adapterList,
		detected:    detected,
		selected:    selected,
		scopes:      scopes,
		scopeIdx:    scopeIdx,
		actions:     actions,
		actionIdx:   0,
		group:       groupPlatforms,
		cursor:      0,
		firstSelect: true,
	}

	out, err := Run(m)
	if err != nil {
		return InstallModalResult{}, err
	}
	fm, ok := out.(installModalModel)
	if !ok {
		return InstallModalResult{}, nil
	}
	return fm.result, nil
}

type scopeOpt struct{ label, value string }
type actionOpt struct{ label, value string }

type modalGroup int

const (
	groupPlatforms modalGroup = iota
	groupScope
	groupAction
)

type installModalModel struct {
	theme    *ui.Theme
	skill    string
	adapters []adapters.Adapter
	detected map[string]bool

	selected  map[string]bool
	scopes    []scopeOpt
	scopeIdx  int
	actions   []actionOpt
	actionIdx int

	group  modalGroup
	cursor int

	width, height int
	firstSelect   bool
	done          bool
	result        InstallModalResult
}

func (m installModalModel) Init() tea.Cmd { return nil }

func (m installModalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		k := msg.String()
		switch k {
		case "ctrl+c", "q":
			m.done = true
			return m, tea.Quit
		case "esc":
			m.done = true
			return m, tea.Quit
		case "tab":
			return m.nextGroup(1), nil
		case "shift+tab":
			return m.nextGroup(-1), nil
		case "up", "k":
			m.cursor = clamp(m.cursor-1, 0, m.groupLen()-1)
			m.syncGroupIndex()
			return m, nil
		case "down", "j":
			m.cursor = clamp(m.cursor+1, 0, m.groupLen()-1)
			m.syncGroupIndex()
			return m, nil
		case " ":
			if m.group == groupPlatforms {
				m = m.togglePlatform()
			}
			return m, nil
		case "enter":
			return m.onEnter()
		}
	}
	return m, nil
}

func (m installModalModel) nextGroup(dir int) installModalModel {
	m.group = modalGroup((int(m.group) + dir + 3) % 3)
	switch m.group {
	case groupPlatforms:
		m.cursor = 0
	case groupScope:
		m.cursor = m.scopeIdx
	case groupAction:
		m.cursor = m.actionIdx
	}
	return m
}

func (m *installModalModel) syncGroupIndex() {
	switch m.group {
	case groupScope:
		m.scopeIdx = m.cursor
	case groupAction:
		m.actionIdx = m.cursor
	}
}

func (m installModalModel) togglePlatform() installModalModel {
	if m.cursor < 0 || m.cursor >= len(m.adapters) {
		return m
	}
	name := m.adapters[m.cursor].Name
	if m.selected[name] {
		delete(m.selected, name)
	} else {
		m.selected[name] = true
	}
	return m
}

func (m installModalModel) onEnter() (tea.Model, tea.Cmd) {
	switch m.group {
	case groupPlatforms:
		// Enter toggles in the platforms group, matching `space`.
		return m.togglePlatform(), nil
	case groupScope:
		m.scopeIdx = m.cursor
		return m.nextGroup(1), nil
	case groupAction:
		m.actionIdx = m.cursor
		return m.commit()
	}
	return m, nil
}

func (m installModalModel) commit() (tea.Model, tea.Cmd) {
	sel := m.actions[m.actionIdx].value
	switch sel {
	case "profile":
		m.result = InstallModalResult{EditProfile: true}
	case "cancel":
		m.result = InstallModalResult{}
	default:
		plats := make([]string, 0, len(m.selected))
		for _, a := range m.adapters {
			if m.selected[a.Name] {
				plats = append(plats, a.Name)
			}
		}
		m.result = InstallModalResult{
			Platforms: plats,
			Scope:     m.scopes[m.scopeIdx].value,
			Confirmed: true,
		}
	}
	m.done = true
	return m, tea.Quit
}

func (m installModalModel) groupLen() int {
	switch m.group {
	case groupPlatforms:
		return len(m.adapters)
	case groupScope:
		return len(m.scopes)
	case groupAction:
		return len(m.actions)
	}
	return 0
}

func (m installModalModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	th := m.theme

	header := Header(th, HeaderSpec{
		Version: versionString,
		Section: "Install",
		Meta:    th.Meta.Render("skill: ") + th.Brand.Render(m.skill),
	}, m.width)

	body := m.renderBody()
	focused := th.Meta.Render("focused: ") + th.Brand.Render(m.groupLabel())
	footer := Footer(th, m.hints(), focused, m.width)

	bodyH := m.height - lipgloss.Height(header) - lipgloss.Height(footer) - 1
	if bodyH < 5 {
		bodyH = 5
	}
	return Frame(header, body, footer, bodyH)
}

func (m installModalModel) groupLabel() string {
	switch m.group {
	case groupPlatforms:
		return "platforms"
	case groupScope:
		return "scope"
	case groupAction:
		return "action"
	}
	return ""
}

func (m installModalModel) renderBody() string {
	th := m.theme
	width := m.width - 4
	if width < 40 {
		width = 40
	}

	var sb strings.Builder
	sb.WriteString("  ")
	sb.WriteString(th.DetailTitle.Render("Install " + m.skill + " to:"))
	sb.WriteString("\n\n")

	sb.WriteString(m.renderGroup(groupPlatforms, "PLATFORMS", m.platformRows(), width))
	sb.WriteString("\n")
	sb.WriteString(m.renderGroup(groupScope, "SCOPE", m.scopeRows(), width))
	sb.WriteString("\n")
	sb.WriteString(m.renderGroup(groupAction, "ACTION", m.actionRows(), width))

	return sb.String()
}

func (m installModalModel) renderGroup(g modalGroup, label string, rows []string, width int) string {
	th := m.theme
	var sb strings.Builder
	titleStyle := th.SectionTitle
	if m.group == g {
		titleStyle = th.Brand
	}
	sb.WriteString("  " + titleStyle.Render(label) + "\n")
	for _, r := range rows {
		sb.WriteString("  " + r + "\n")
	}
	_ = width
	return sb.String()
}

func (m installModalModel) platformRows() []string {
	th := m.theme
	rows := make([]string, 0, len(m.adapters))
	for i, a := range m.adapters {
		box := "[ ]"
		if m.selected[a.Name] {
			box = "[✓]"
		}
		label := a.Name
		if m.detected[a.Name] {
			label += "  " + th.Detail.Render("(detected)")
		} else {
			label += "  " + th.RowDim.Render("(not detected)")
		}
		cursorHere := m.group == groupPlatforms && i == m.cursor
		var line string
		switch {
		case cursorHere:
			line = th.Bullet.Render("▸") + " " + th.RowSelected.Render(box+" "+a.Name) + "  " + label[len(a.Name)+2:]
		case m.selected[a.Name]:
			line = "  " + th.Success.Render(box) + " " + th.RowUnselected.Render(a.Name) + "  " + label[len(a.Name)+2:]
		default:
			line = "  " + th.RowDim.Render(box) + " " + th.RowUnselected.Render(a.Name) + "  " + label[len(a.Name)+2:]
		}
		rows = append(rows, line)
	}
	return rows
}

func (m installModalModel) scopeRows() []string {
	th := m.theme
	rows := make([]string, 0, len(m.scopes))
	for i, s := range m.scopes {
		cursorHere := m.group == groupScope && i == m.cursor
		isCurrent := i == m.scopeIdx
		marker := "( )"
		if isCurrent {
			marker = "(●)"
		}
		var line string
		switch {
		case cursorHere:
			line = th.Bullet.Render("▸") + " " + th.RowSelected.Render(marker+" "+s.label)
		case isCurrent:
			line = "  " + th.Success.Render(marker) + " " + th.RowUnselected.Render(s.label)
		default:
			line = "  " + th.RowDim.Render(marker) + " " + th.RowUnselected.Render(s.label)
		}
		rows = append(rows, line)
	}
	return rows
}

func (m installModalModel) actionRows() []string {
	th := m.theme
	rows := make([]string, 0, len(m.actions))
	for i, a := range m.actions {
		cursorHere := m.group == groupAction && i == m.cursor
		var line string
		if cursorHere {
			line = th.Bullet.Render("▸") + " " + th.RowSelected.Render(a.label)
		} else {
			line = "  " + th.RowUnselected.Render(a.label)
		}
		rows = append(rows, line)
	}
	return rows
}

func (m installModalModel) hints() []KeyHint {
	base := []KeyHint{
		{Keys: "↑↓", Label: "select"},
		{Keys: "tab", Label: "next group"},
	}
	switch m.group {
	case groupPlatforms:
		base = append(base, KeyHint{Keys: "space", Label: "toggle"})
	}
	base = append(base,
		KeyHint{Keys: "↵", Label: "confirm"},
		KeyHint{Keys: "esc", Label: "back"},
	)
	return base
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
