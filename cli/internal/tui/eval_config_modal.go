package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// EvalConfigChoices bundles what the Config modal returns to the caller.
type EvalConfigChoices struct {
	Arms      []string
	Scenarios []string
	Runner    string
	Runs      int
	Parallel  int
	Confirmed bool
}

// EvalConfigInputs is what the caller feeds the modal.
type EvalConfigInputs struct {
	Skill      string
	Arms       []string            // all configured; preselected = all
	Scenarios  []EvalScenarioEntry // preselected = all
	Runners    []EvalRunnerEntry   // with availability + default suggestion
	Runs       []int               // menu: 1, 3, 5, 10
	Parallel   []int               // menu: 1, 2, 4, 8
	DefaultRun int
	DefaultPar int
}

// EvalScenarioEntry is one row in the scenarios group.
type EvalScenarioEntry struct {
	ID       string
	Family   string
	Sessions int
}

// EvalRunnerEntry is one row in the runner group.
type EvalRunnerEntry struct {
	Name      string
	Available bool
	Version   string
	Reason    string
}

// RunEvalConfigModal opens the config modal and returns the chosen
// configuration. Confirmed=false on cancel.
func RunEvalConfigModal(theme *ui.Theme, crumb string, in EvalConfigInputs) (EvalConfigChoices, error) {
	m := evalConfigModel{
		theme:        theme,
		crumb:        crumb,
		in:           in,
		armSel:       preselectAll(len(in.Arms)),
		scenarioSel:  preselectAll(len(in.Scenarios)),
		runnerIdx:    indexOfDefaultRunner(in.Runners),
		runIdx:       indexOfInt(in.Runs, in.DefaultRun, 0),
		parIdx:       indexOfInt(in.Parallel, in.DefaultPar, 0),
	}
	out, err := Run(m)
	if err != nil {
		return EvalConfigChoices{}, err
	}
	em, ok := out.(evalConfigModel)
	if !ok {
		return EvalConfigChoices{}, nil
	}
	return em.result(), nil
}

// --- model ------------------------------------------------------------------

type cfgGroup int

const (
	grpArms cfgGroup = iota
	grpScenarios
	grpRunner
	grpRuns
	grpParallel
	grpAction
)

type evalConfigModel struct {
	theme         *ui.Theme
	crumb         string
	in            EvalConfigInputs
	armSel        []bool
	scenarioSel   []bool
	runnerIdx     int
	runIdx        int
	parIdx        int
	group         cfgGroup
	cursor        int
	actionIdx     int // 0 = Run, 1 = Cancel
	width, height int
	done          bool
	confirmed     bool
}

func (m evalConfigModel) Init() tea.Cmd { return nil }

func (m evalConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		keys := DefaultKeys()
		switch {
		case key.Matches(msg, keys.Quit), key.Matches(msg, keys.Back):
			m.done = true
			return m, tea.Quit
		case key.Matches(msg, keys.Up):
			m.cursor = clampInt(m.cursor-1, 0, m.groupLen()-1)
			return m, nil
		case key.Matches(msg, keys.Down):
			m.cursor = clampInt(m.cursor+1, 0, m.groupLen()-1)
			return m, nil
		case key.Matches(msg, keys.Enter):
			return m.onEnter()
		}
		// Group cycling and multi-select toggle aren't shared keymap verbs.
		switch msg.String() {
		case "tab":
			m.group = cfgGroup((int(m.group) + 1) % 6)
			m.cursor = m.cursorDefault()
			return m, nil
		case "shift+tab":
			m.group = cfgGroup((int(m.group) + 5) % 6)
			m.cursor = m.cursorDefault()
			return m, nil
		case " ":
			m.toggleCurrent()
			return m, nil
		}
	}
	return m, nil
}

func (m *evalConfigModel) toggleCurrent() {
	switch m.group {
	case grpArms:
		if m.cursor < len(m.armSel) {
			m.armSel[m.cursor] = !m.armSel[m.cursor]
		}
	case grpScenarios:
		if m.cursor < len(m.scenarioSel) {
			m.scenarioSel[m.cursor] = !m.scenarioSel[m.cursor]
		}
	case grpRunner:
		m.runnerIdx = m.cursor
	case grpRuns:
		m.runIdx = m.cursor
	case grpParallel:
		m.parIdx = m.cursor
	}
}

func (m evalConfigModel) onEnter() (tea.Model, tea.Cmd) {
	if m.group != grpAction {
		m.toggleCurrent()
		return m, nil
	}
	if m.cursor == 0 {
		m.confirmed = true
	}
	m.done = true
	return m, tea.Quit
}

func (m evalConfigModel) cursorDefault() int {
	switch m.group {
	case grpRunner:
		return m.runnerIdx
	case grpRuns:
		return m.runIdx
	case grpParallel:
		return m.parIdx
	case grpAction:
		return 0
	}
	return 0
}

func (m evalConfigModel) groupLen() int {
	switch m.group {
	case grpArms:
		return len(m.in.Arms)
	case grpScenarios:
		return len(m.in.Scenarios)
	case grpRunner:
		return len(m.in.Runners)
	case grpRuns:
		return len(m.in.Runs)
	case grpParallel:
		return len(m.in.Parallel)
	case grpAction:
		return 2
	}
	return 0
}

func (m evalConfigModel) result() EvalConfigChoices {
	r := EvalConfigChoices{Confirmed: m.confirmed}
	for i, on := range m.armSel {
		if on {
			r.Arms = append(r.Arms, m.in.Arms[i])
		}
	}
	for i, on := range m.scenarioSel {
		if on {
			r.Scenarios = append(r.Scenarios, m.in.Scenarios[i].ID)
		}
	}
	if m.runnerIdx >= 0 && m.runnerIdx < len(m.in.Runners) {
		r.Runner = m.in.Runners[m.runnerIdx].Name
	}
	if m.runIdx >= 0 && m.runIdx < len(m.in.Runs) {
		r.Runs = m.in.Runs[m.runIdx]
	}
	if m.parIdx >= 0 && m.parIdx < len(m.in.Parallel) {
		r.Parallel = m.in.Parallel[m.parIdx]
	}
	return r
}

// --- view -------------------------------------------------------------------

func (m evalConfigModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	th := m.theme
	header := Header(th, HeaderSpec{
		Version: versionString,
		Section: m.crumb,
		Meta:    th.Meta.Render("skill: ") + th.Brand.Render(m.in.Skill),
	}, m.width)

	var sb strings.Builder
	sb.WriteString("  " + th.DetailTitle.Render("Configure eval run") + "\n\n")
	sb.WriteString(m.renderArms())
	sb.WriteString("\n")
	sb.WriteString(m.renderScenarios())
	sb.WriteString("\n")
	sb.WriteString(m.renderRunner())
	sb.WriteString("\n")
	sb.WriteString(m.renderRuns())
	sb.WriteString("\n")
	sb.WriteString(m.renderParallel())
	sb.WriteString("\n")
	sb.WriteString(m.renderAction())

	footer := Footer(th, []KeyHint{
		{Keys: "tab", Label: "next group"},
		{Keys: "↑↓", Label: "select"},
		{Keys: "space", Label: "toggle"},
		{Keys: "↵", Label: "confirm"},
		{Keys: "esc", Label: "back"},
	}, th.Meta.Render("focus: ")+th.Brand.Render(groupLabel(m.group)), m.width)

	bodyH := m.height - lipgloss.Height(header) - lipgloss.Height(footer) - 1
	if bodyH < 10 {
		bodyH = 10
	}
	return Frame(header, sb.String(), footer, bodyH)
}

func (m evalConfigModel) renderArms() string {
	return m.renderMultiGroup(grpArms, "ARMS", m.in.Arms, func(i int) (string, string, bool) {
		return m.in.Arms[i], "", m.armSel[i]
	})
}

func (m evalConfigModel) renderScenarios() string {
	return m.renderMultiGroup(grpScenarios, "SCENARIOS",
		make([]string, len(m.in.Scenarios)), // unused slice for row count
		func(i int) (string, string, bool) {
			s := m.in.Scenarios[i]
			sub := fmt.Sprintf("%s · %d sessions", s.Family, s.Sessions)
			return s.ID, sub, m.scenarioSel[i]
		})
}

func (m evalConfigModel) renderRunner() string {
	return m.renderRadioGroup(grpRunner, "RUNNER", len(m.in.Runners), m.runnerIdx,
		func(i int) (string, string) {
			r := m.in.Runners[i]
			note := r.Version
			if !r.Available {
				note = r.Reason
			}
			return r.Name, note
		})
}

func (m evalConfigModel) renderRuns() string {
	return m.renderRadioGroup(grpRuns, "RUNS PER CONFIG", len(m.in.Runs), m.runIdx,
		func(i int) (string, string) {
			return fmt.Sprintf("%d", m.in.Runs[i]), ""
		})
}

func (m evalConfigModel) renderParallel() string {
	return m.renderRadioGroup(grpParallel, "PARALLEL", len(m.in.Parallel), m.parIdx,
		func(i int) (string, string) {
			return fmt.Sprintf("%d", m.in.Parallel[i]), ""
		})
}

func (m evalConfigModel) renderAction() string {
	th := m.theme
	var sb strings.Builder
	titleStyle := th.SectionTitle
	if m.group == grpAction {
		titleStyle = th.Brand
	}
	sb.WriteString("  " + titleStyle.Render("ACTION") + "\n")
	opts := []string{"[Run]", "[Cancel]"}
	for i, opt := range opts {
		cursorHere := m.group == grpAction && i == m.cursor
		styled := th.RowUnselected.Render(opt)
		if cursorHere {
			styled = th.RowSelected.Render(opt)
		}
		prefix := "  "
		if cursorHere {
			prefix = th.Bullet.Render("▸") + " "
		}
		sb.WriteString(prefix + styled + "  ")
	}
	sb.WriteString("\n")
	return sb.String()
}

func (m evalConfigModel) renderMultiGroup(g cfgGroup, label string, rows []string, fn func(int) (string, string, bool)) string {
	th := m.theme
	var sb strings.Builder
	titleStyle := th.SectionTitle
	if m.group == g {
		titleStyle = th.Brand
	}
	sb.WriteString("  " + titleStyle.Render(label) + "\n")
	for i := range rows {
		name, sub, on := fn(i)
		cursorHere := m.group == g && i == m.cursor
		box := "[ ]"
		if on {
			box = "[✓]"
		}
		var line string
		switch {
		case cursorHere:
			line = th.Bullet.Render("▸") + " " + th.RowSelected.Render(box+" "+name)
		case on:
			line = "  " + th.Success.Render(box) + " " + th.RowUnselected.Render(name)
		default:
			line = "  " + th.RowDim.Render(box) + " " + th.RowUnselected.Render(name)
		}
		if sub != "" {
			line += "  " + th.Detail.Render(sub)
		}
		sb.WriteString(line + "\n")
	}
	return sb.String()
}

func (m evalConfigModel) renderRadioGroup(g cfgGroup, label string, count, current int, fn func(int) (string, string)) string {
	th := m.theme
	var sb strings.Builder
	titleStyle := th.SectionTitle
	if m.group == g {
		titleStyle = th.Brand
	}
	sb.WriteString("  " + titleStyle.Render(label) + "  ")
	for i := 0; i < count; i++ {
		name, note := fn(i)
		cursorHere := m.group == g && i == m.cursor
		isCurrent := i == current
		marker := "( )"
		if isCurrent {
			marker = "(●)"
		}
		var styled string
		switch {
		case cursorHere:
			styled = th.RowSelected.Render(marker + " " + name)
		case isCurrent:
			styled = th.Success.Render(marker) + " " + th.RowUnselected.Render(name)
		default:
			styled = th.RowDim.Render(marker) + " " + th.RowUnselected.Render(name)
		}
		sb.WriteString(styled)
		if note != "" {
			sb.WriteString(" " + th.Detail.Render(note))
		}
		sb.WriteString("    ")
	}
	sb.WriteString("\n")
	return sb.String()
}

// --- helpers ----------------------------------------------------------------

func preselectAll(n int) []bool {
	out := make([]bool, n)
	for i := range out {
		out[i] = true
	}
	return out
}

func indexOfDefaultRunner(rs []EvalRunnerEntry) int {
	for i, r := range rs {
		if r.Available {
			return i
		}
	}
	return 0
}

func indexOfInt(xs []int, v, fallback int) int {
	for i, x := range xs {
		if x == v {
			return i
		}
	}
	return fallback
}

func groupLabel(g cfgGroup) string {
	switch g {
	case grpArms:
		return "arms"
	case grpScenarios:
		return "scenarios"
	case grpRunner:
		return "runner"
	case grpRuns:
		return "runs"
	case grpParallel:
		return "parallel"
	case grpAction:
		return "action"
	}
	return ""
}
