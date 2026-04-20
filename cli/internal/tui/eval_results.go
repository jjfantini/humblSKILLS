package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// EvalResultsData is the payload the caller hands the Results screen.
type EvalResultsData struct {
	Skill       string
	Iteration   int
	ReportHTML  string
	Sparklines  map[string]EvalResultsSeries // keyed by arm
	CrossTable  [][]string                   // rows of (arm, pass@1, pass@K, Δ)
}

// EvalResultsSeries is one arm's series for the ASCII sparklines.
type EvalResultsSeries struct {
	PassRate []float64
	Tokens   []float64
	PassAtK  []float64
	Wiki     []float64
}

// RunEvalResults displays the post-run sparkline summary. onAction is
// invoked when the user picks an action ("open-html", "markdown", "rerun").
func RunEvalResults(theme *ui.Theme, crumb string, data EvalResultsData, onAction func(string) error) error {
	m := evalResultsModel{theme: theme, crumb: crumb, data: data}
	out, err := Run(m)
	if err != nil {
		return err
	}
	if em, ok := out.(evalResultsModel); ok && em.action != "" {
		return onAction(em.action)
	}
	return nil
}

// --- model ------------------------------------------------------------------

type evalResultsModel struct {
	theme         *ui.Theme
	crumb         string
	data          EvalResultsData
	width, height int
	action        string
	done          bool
}

func (m evalResultsModel) Init() tea.Cmd { return nil }

func (m evalResultsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.done = true
			return m, tea.Quit
		case "h":
			m.action = "open-html"
			m.done = true
			return m, tea.Quit
		case "m":
			m.action = "markdown"
			m.done = true
			return m, tea.Quit
		case "r":
			m.action = "rerun"
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m evalResultsModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	th := m.theme
	header := Header(th, HeaderSpec{
		Version: versionString,
		Section: m.crumb,
		Meta: fmt.Sprintf("%s · iter %d",
			th.Brand.Render(m.data.Skill),
			m.data.Iteration),
	}, m.width)

	var sb strings.Builder
	arms := make([]string, 0, len(m.data.Sparklines))
	for a := range m.data.Sparklines {
		arms = append(arms, a)
	}
	sort.Strings(arms)

	sb.WriteString("  " + th.SectionTitle.Render("PASS_RATE") + "\n")
	for _, a := range arms {
		s := m.data.Sparklines[a]
		sb.WriteString(fmt.Sprintf("  %-12s %s\n", a, sparkline(s.PassRate)))
	}
	sb.WriteString("\n")

	sb.WriteString("  " + th.SectionTitle.Render("TOKENS") + "\n")
	for _, a := range arms {
		s := m.data.Sparklines[a]
		sb.WriteString(fmt.Sprintf("  %-12s %s\n", a, sparkline(s.Tokens)))
	}
	sb.WriteString("\n")

	sb.WriteString("  " + th.SectionTitle.Render("PASS^K") + "\n")
	for _, a := range arms {
		s := m.data.Sparklines[a]
		sb.WriteString(fmt.Sprintf("  %-12s %s\n", a, sparkline(s.PassAtK)))
	}
	sb.WriteString("\n")

	if len(m.data.CrossTable) > 0 {
		sb.WriteString("  " + th.SectionTitle.Render("CROSS-SECTION") + "\n")
		for _, row := range m.data.CrossTable {
			sb.WriteString("  " + strings.Join(row, "  ") + "\n")
		}
	}

	footer := Footer(th, []KeyHint{
		{Keys: "h", Label: "open html"},
		{Keys: "m", Label: "markdown"},
		{Keys: "r", Label: "re-run"},
		{Keys: "esc", Label: "back"},
	}, "", m.width)
	bodyH := m.height - lipgloss.Height(header) - lipgloss.Height(footer) - 1
	if bodyH < 5 {
		bodyH = 5
	}
	return Frame(header, sb.String(), footer, bodyH)
}

// sparkline reimplements the one in report/ to keep the TUI package free
// of a report dep. Uses the same character set so output reads identically.
func sparkline(vals []float64) string {
	chars := []rune("▁▂▃▄▅▆▇█")
	if len(vals) == 0 {
		return ""
	}
	min, max := vals[0], vals[0]
	for _, v := range vals {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	rng := max - min
	var b strings.Builder
	for _, v := range vals {
		i := 0
		if rng > 0 {
			i = int(((v-min)/rng)*float64(len(chars)-1) + 0.5)
			if i < 0 {
				i = 0
			}
			if i > len(chars)-1 {
				i = len(chars) - 1
			}
		}
		b.WriteRune(chars[i])
	}
	return b.String()
}
