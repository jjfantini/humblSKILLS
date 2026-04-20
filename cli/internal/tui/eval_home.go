package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// EvalHomeItem is one row in the Eval Home screen's left pane.
type EvalHomeItem struct {
	Skill          string
	SkillDir       string
	HasEvals       bool
	ScenarioCount  int
	Configurations []string
	IterationCount int
	LastRun        time.Time
	LastPassRates  map[string]float64
	Runner         string
}

// EvalHomeData is the payload handed to the Home screen.
type EvalHomeData struct {
	Workspace string
	Items     []EvalHomeItem
}

// RunEvalHome opens the Eval Home screen. onRun launches a run for the
// chosen skill and re-enters the Home screen when it returns.
//
// Return values: no value; the function blocks until the user quits with
// ESC or a run is complete.
func RunEvalHome(theme *ui.Theme, crumb string, load func() (EvalHomeData, error), onRun func(string) error) error {
	for {
		data, err := load()
		if err != nil {
			return err
		}
		m := evalHomeModel{theme: theme, crumb: crumb, data: data}
		out, err := Run(m)
		if err != nil {
			return err
		}
		em, ok := out.(evalHomeModel)
		if !ok {
			return nil
		}
		if em.quit {
			return nil
		}
		if em.chosenRun != "" {
			if err := onRun(em.chosenRun); err != nil {
				return err
			}
			// Re-enter.
			continue
		}
		return nil
	}
}

// RunEvalHomeOr is a thin wrapper that falls back to onRun if the TUI is
// unavailable. Used by the CLI entrypoint.
func RunEvalHomeOr(theme *ui.Theme, crumb string, load func() (EvalHomeData, error), onRun func(string) error) error {
	return RunEvalHome(theme, crumb, load, onRun)
}

// --- model ------------------------------------------------------------------

type evalHomeModel struct {
	theme         *ui.Theme
	crumb         string
	data          EvalHomeData
	cursor        int
	width, height int
	quit          bool
	chosenRun     string
}

func (m evalHomeModel) Init() tea.Cmd { return nil }

func (m evalHomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quit = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.data.Items)-1 {
				m.cursor++
			}
		case "enter", "r":
			if len(m.data.Items) > 0 {
				m.chosenRun = m.data.Items[m.cursor].Skill
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m evalHomeModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	th := m.theme
	header := Header(th, HeaderSpec{
		Version: versionString,
		Section: m.crumb,
		Meta:    th.Meta.Render("workspace: ") + th.Name.Render(abbreviatePath(m.data.Workspace, 40)),
	}, m.width)
	footer := Footer(th, []KeyHint{
		{Keys: "↑↓", Label: "select"},
		{Keys: "↵/r", Label: "run"},
		{Keys: "esc", Label: "back"},
	}, "", m.width)
	bodyH := m.height - lipgloss.Height(header) - lipgloss.Height(footer) - 1
	if bodyH < 5 {
		bodyH = 5
	}
	return Frame(header, m.renderBody(), footer, bodyH)
}

func (m evalHomeModel) renderBody() string {
	th := m.theme
	if len(m.data.Items) == 0 {
		return "  " + th.Detail.Render("no skills with evals/ found — try `humblskills eval init <skill>`")
	}
	leftW, rightW := m.paneWidths()
	left := m.renderLeft(leftW)
	right := m.renderRight(rightW)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m evalHomeModel) paneWidths() (int, int) {
	left := 36
	if cap := m.width / 3; cap > 0 && left > cap {
		left = cap
	}
	if left < 24 {
		left = 24
	}
	return left, m.width - left
}

func (m evalHomeModel) renderLeft(width int) string {
	th := m.theme
	var sb strings.Builder
	sb.WriteString("  " + th.SectionTitle.Render("SKILLS (evals/)") + "\n\n")
	for i, it := range m.data.Items {
		prefix := "  "
		if i == m.cursor {
			prefix = th.Bullet.Render("▌") + " "
		}
		dot := th.DotNo.Render("○")
		if it.HasEvals {
			dot = th.DotOK.Render("●")
		}
		name := it.Skill
		nameStyle := th.RowUnselected
		if i == m.cursor {
			nameStyle = th.RowSelected
		}
		stat := th.Crumb.Render("no evals")
		if it.HasEvals {
			stat = th.Crumb.Render(fmt.Sprintf("iter %d", it.IterationCount))
		}
		line := prefix + dot + " " + nameStyle.Render(name)
		row := padLine(line, width-len(stat)-2) + stat
		sb.WriteString(row + "\n")
	}
	return sb.String()
}

func (m evalHomeModel) renderRight(width int) string {
	th := m.theme
	if m.cursor < 0 || m.cursor >= len(m.data.Items) {
		return ""
	}
	it := m.data.Items[m.cursor]
	var sb strings.Builder
	bar := th.Divider.Render("│")
	sb.WriteString(bar + " " + th.DetailTitle.Render(it.Skill) + "\n")
	sb.WriteString(bar + "\n")

	if !it.HasEvals {
		sb.WriteString(bar + " " + th.Detail.Render("No evals configured.") + "\n")
		sb.WriteString(bar + " " + th.Detail.Render("Run `humblskills eval init "+it.Skill+"`") + "\n")
		return sb.String()
	}
	sb.WriteString(bar + " " + kv(th, "scenarios", fmt.Sprintf("%d", it.ScenarioCount)))
	sb.WriteString(bar + " " + kv(th, "arms", strings.Join(it.Configurations, ", ")))
	if it.IterationCount > 0 {
		age := time.Since(it.LastRun).Truncate(time.Minute)
		sb.WriteString(bar + " " + kv(th, "last run", fmt.Sprintf("%s ago (iter %d)", ageString(age), it.IterationCount)))
		if it.Runner != "" {
			sb.WriteString(bar + " " + kv(th, "runner", it.Runner))
		}
		keys := sortedArms(it.LastPassRates)
		if len(keys) > 0 {
			var parts []string
			for _, a := range keys {
				parts = append(parts, fmt.Sprintf("%s %.2f", shortArm(a), it.LastPassRates[a]))
			}
			sb.WriteString(bar + " " + kv(th, "last pass_rate", strings.Join(parts, " · ")))
		}
	} else {
		sb.WriteString(bar + " " + th.Detail.Render("no runs yet - press r / enter to kick off the first one") + "\n")
	}
	sb.WriteString(bar + "\n")
	sb.WriteString(bar + " " + th.Detail.Render("[r] run  [enter] run") + "\n")
	_ = width
	return sb.String()
}

func kv(th *ui.Theme, k, v string) string {
	return th.KVKey.Render(padRightStr(k, 14)) + "  " + th.KVValue.Render(v) + "\n"
}

func padRightStr(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

func padLine(s string, w int) string {
	cw := lipgloss.Width(s)
	if cw >= w {
		return s
	}
	return s + strings.Repeat(" ", w-cw)
}

func ageString(d time.Duration) string {
	if d < time.Minute {
		return "moments"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

func shortArm(a string) string {
	switch a {
	case "smart_skill":
		return "smart"
	case "flat_skill":
		return "flat"
	case "no_skill":
		return "none"
	}
	return a
}

func sortedArms(m map[string]float64) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func abbreviatePath(p string, max int) string {
	if len(p) <= max {
		return p
	}
	return "…" + p[len(p)-max+1:]
}
