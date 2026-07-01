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
// successfully: one headline plus optional key/value detail lines.
type StatusResult struct {
	Headline string
	Lines    []string
}

// statusDoneMsg carries fn's result back into the bubbletea event loop.
type statusDoneMsg struct {
	result StatusResult
	err    error
}

// StatusModel renders a single blocking operation (e.g. "refresh the
// registry") as a spinner while it runs, then a persistent success/error
// screen that stays up until the user dismisses it or - on success only -
// an optional auto-return timer fires. This mirrors ProgressModel's
// done-screen behavior for operations that don't have granular per-target
// progress to report, just one result at the end (registry refresh today).
type StatusModel struct {
	theme   *ui.Theme
	command string // header breadcrumb, e.g. "registry"
	label   string // spinner label while running, e.g. "refreshing registry…"
	spin    spinner.Model
	fn      func() (StatusResult, error)

	running bool
	result  StatusResult
	err     error

	autoReturn autoReturnTimer
	width      int
}

// NewStatusModel builds a model that runs fn once fn is kicked off by Init.
// autoReturn is the countdown duration for the success screen (0 disables
// it, matching Profile.StatusAutoReturnDuration's semantics).
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
		autoReturn: autoReturnTimer{duration: autoReturn},
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
		m.width = msg.Width
		return m, nil

	case statusDoneMsg:
		m.running = false
		m.result = msg.result
		m.err = msg.err
		if m.err == nil {
			return m, m.autoReturn.Start()
		}
		return m, nil

	case autoReturnTickMsg:
		quit, cmd := m.autoReturn.Tick()
		if quit {
			return m, tea.Quit
		}
		return m, cmd

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if !m.running && (msg.String() == "q" || msg.String() == "enter" || msg.String() == "esc") {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m StatusModel) View() string {
	th := m.theme
	header := Header(th, HeaderSpec{Section: m.command}, m.width)

	var sb strings.Builder
	switch {
	case m.running:
		sb.WriteString("  " + m.spin.View() + " " + th.Crumb.Render(m.label) + "\n")
	case m.err != nil:
		sb.WriteString("  " + th.Error.Render("✗ ") + th.Name.Render("failed") + "\n\n")
		sb.WriteString("  " + th.Error.Render("error: ") + th.Detail.Render(m.err.Error()) + "\n")
	default:
		headline := m.result.Headline
		if headline == "" {
			headline = "done"
		}
		sb.WriteString("  " + th.Success.Render("✓ ") + th.Name.Render(headline) + "\n")
		for _, line := range m.result.Lines {
			sb.WriteString("  " + th.Detail.Render(line) + "\n")
		}
	}

	var footer string
	switch {
	case m.running:
		footer = Footer(th, []KeyHint{{Keys: "ctrl+c", Label: "abort"}}, "", m.width)
	case m.autoReturn.Active():
		footer = Footer(th, []KeyHint{{Keys: "enter/q", Label: "close now"}},
			th.Detail.Render(fmt.Sprintf("closing in %ds", m.autoReturn.RemainingSeconds())), m.width)
	default:
		footer = Footer(th, []KeyHint{{Keys: "enter/q", Label: "close"}}, "", m.width)
	}
	return header + "\n\n" + sb.String() + "\n" + footer
}

// ExecuteWithStatus runs fn in the background while a StatusModel shows a
// spinner, then its success/error result, staying on screen until the user
// dismisses it or (success only) autoReturn elapses.
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
