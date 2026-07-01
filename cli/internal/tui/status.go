package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// StatusResult is the summary a StatusModel renders once fn completes
// successfully: one headline plus optional detail content. Exactly one of
// Raw or Lines is normally set: Raw is pre-rendered/already-styled text
// (e.g. text captured verbatim from app.UI calls) that's shown as-is; Lines
// is a list of plain strings the model styles itself with the muted detail
// color. Raw takes precedence when both are set.
type StatusResult struct {
	Headline string
	Lines    []string
	Raw      string
}

// statusDoneMsg carries fn's result back into the bubbletea event loop.
type statusDoneMsg struct {
	result StatusResult
	err    error
}

// StatusModel renders a single blocking operation (e.g. "refresh the
// registry") as a spinner while it runs, then a persistent success/error
// screen that stays up until the user dismisses it or - on success only -
// an optional auto-return timer fires once they've seen the whole result.
// This mirrors ProgressModel's done-screen behavior for operations that
// don't have granular per-target progress to report, just one result (or,
// via StatusResult.Raw, arbitrary captured output) at the end.
type StatusModel struct {
	theme   *ui.Theme
	command string // header breadcrumb, e.g. "registry"
	label   string // spinner label while running, e.g. "refreshing registry…"
	spin    spinner.Model
	fn      func() (StatusResult, error)

	running bool
	result  StatusResult
	err     error
	width   int
	height  int

	// resultView owns the scrollable result body plus the shared
	// auto-return countdown — see scrollstatus.go.
	resultView scrollableDone
}

// NewStatusModel builds a model that runs fn once fn is kicked off by Init.
// autoReturn is the countdown duration for the success screen (0 disables
// it, matching Profile.StatusAutoReturnDuration's semantics); even when
// enabled, the countdown waits for the user to scroll to the bottom if the
// result doesn't already fit on screen — see scrollableDone.ArmIfReady.
func NewStatusModel(theme *ui.Theme, command, label string, autoReturn time.Duration, fn func() (StatusResult, error)) StatusModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = theme.Brand

	return StatusModel{
		theme:      theme,
		command:    command,
		label:      label,
		spin:       sp,
		fn:         fn,
		running:    true,
		resultView: newScrollableDone(autoReturn),
	}
}

// Err returns fn's terminal error, if any. Meaningful only after the model
// has exited.
func (m StatusModel) Err() error { return m.err }

// Result returns fn's result. Meaningful only after the model has exited
// without error.
func (m StatusModel) Result() StatusResult { return m.result }

func (m StatusModel) Init() tea.Cmd {
	return tea.Batch(m.spin.Tick, runStatusFn(m.fn))
}

func runStatusFn(fn func() (StatusResult, error)) tea.Cmd {
	return func() tea.Msg {
		res, err := fn()
		return statusDoneMsg{result: res, err: err}
	}
}

func (m StatusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.refreshResultContent(false)
		return m, nil

	case statusDoneMsg:
		m.running = false
		m.result = msg.result
		m.err = msg.err
		m.refreshResultContent(true)
		return m, m.resultView.ArmIfReady(m.err == nil)

	case autoReturnTickMsg:
		quit, cmd := m.resultView.Tick()
		if quit {
			return m, tea.Quit
		}
		return m, cmd

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case tea.MouseMsg:
		cmd := m.resultView.HandleMouse(msg)
		return m, tea.Batch(cmd, m.resultView.ArmIfReady(!m.running && m.err == nil))

	case tea.KeyMsg:
		if !m.running && (msg.String() == "q" || msg.String() == "enter" || msg.String() == "esc") {
			return m, tea.Quit
		}
		if m.resultView.HandleKey(msg.String()) {
			return m, m.resultView.ArmIfReady(!m.running && m.err == nil)
		}
	}
	return m, nil
}

func (m StatusModel) View() string {
	th := m.theme
	header := Header(th, HeaderSpec{
		Section: m.command,
		Meta:    m.resultView.ScrollIndicator(th),
	}, m.width)

	top := m.renderTop(th)
	footer := m.renderFooter(th)

	if m.running {
		return header + "\n\n" + top + "\n\n" + footer
	}
	return header + "\n\n" + top + "\n\n" + m.resultView.View() + "\n" + footer
}

// renderTop is the fixed, always-visible state line above the scrollable
// body: the spinner while running, or a compact ✓/✗ + short label once done
// (the substantive detail lives in the scrollable body below it).
func (m StatusModel) renderTop(th *ui.Theme) string {
	switch {
	case m.running:
		return "  " + m.spin.View() + " " + th.Crumb.Render(m.label)
	case m.err != nil:
		return "  " + th.Error.Render("✗ ") + th.Name.Render("failed")
	default:
		headline := m.result.Headline
		if headline == "" {
			headline = "done"
		}
		return "  " + th.Success.Render("✓ ") + th.Name.Render(headline)
	}
}

// renderFooter picks the key hints + right-anchored context for the current
// state, matching ProgressModel.renderFooter's pattern.
func (m StatusModel) renderFooter(th *ui.Theme) string {
	switch {
	case m.running:
		return Footer(th, []KeyHint{{Keys: "ctrl+c", Label: "abort"}}, "", m.width)
	case m.err != nil:
		hints := []KeyHint{{Keys: "enter/q", Label: "close"}}
		if m.resultView.Overflows() {
			hints = append(hints, KeyHint{Keys: "↑↓", Label: "scroll"})
		}
		return Footer(th, hints, "", m.width)
	case m.resultView.Active():
		return Footer(th, []KeyHint{{Keys: "enter/q", Label: "close now"}},
			th.Detail.Render(fmt.Sprintf("closing in %ds", m.resultView.RemainingSeconds())), m.width)
	case m.resultView.Enabled():
		return Footer(th, []KeyHint{{Keys: "enter/q", Label: "close"}, {Keys: "↓/end", Label: "scroll to bottom"}},
			th.Detail.Render("scroll to bottom to auto-close"), m.width)
	default:
		return Footer(th, []KeyHint{{Keys: "enter/q", Label: "close"}}, "", m.width)
	}
}

// resizeResultView computes the scrollable body's height budget: total
// height minus the header, the fixed top status line, the blank separator
// rows, and the footer. Mirrors ProgressModel.resizeResultView.
func (m *StatusModel) resizeResultView() {
	if m.width == 0 || m.height == 0 {
		return
	}
	const (
		headerH   = 2
		topH      = 1
		footerH   = 2
		blankRows = 3 // header/top, top/body, body/footer separators
	)
	bodyH := m.height - (headerH + topH + footerH + blankRows)
	if bodyH < 3 {
		bodyH = 3
	}
	m.resultView.Resize(m.width, bodyH)
}

// refreshResultContent recomputes the scrollable body's dimensions and
// content. resetToTop should be true exactly once — when fn finishes — so
// the user reads the result from the start.
func (m *StatusModel) refreshResultContent(resetToTop bool) {
	m.resizeResultView()
	m.resultView.SetContent(m.resultBody(), resetToTop)
}

// resultBody composes the scrollable body: the pre-rendered Raw text if
// set, otherwise Lines styled with the muted detail color, followed by the
// error line (if any) so a failure still shows whatever context was
// captured before it — e.g. sync's "skipping unknown skill" warnings ahead
// of a later hard failure.
func (m StatusModel) resultBody() string {
	th := m.theme
	var sb strings.Builder
	switch {
	case m.result.Raw != "":
		sb.WriteString(strings.TrimRight(m.result.Raw, "\n"))
	case len(m.result.Lines) > 0:
		for i, line := range m.result.Lines {
			if i > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString("  " + th.Detail.Render(line))
		}
	}
	if m.err != nil {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString("  " + th.Error.Render("error: ") + th.Detail.Render(m.err.Error()))
	}
	return sb.String()
}

// ExecuteWithStatus runs fn in the background while a StatusModel shows a
// spinner, then its success/error result, staying on screen until the user
// dismisses it or (success only, once they've seen the whole result)
// autoReturn elapses.
func ExecuteWithStatus(theme *ui.Theme, command, label string, autoReturn time.Duration, fn func() (StatusResult, error)) (StatusResult, error) {
	m, err := Run(NewStatusModel(theme, command, label, autoReturn, fn))
	if err != nil {
		return StatusResult{}, err
	}
	sm, ok := m.(StatusModel)
	if !ok {
		return StatusResult{}, nil
	}
	return sm.result, sm.err
}
