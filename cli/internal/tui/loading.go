package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// loadingDoneMsg carries fn's result back into the bubbletea event loop.
type loadingDoneMsg[T any] struct {
	result T
	err    error
}

// loadingModel is a minimal bubbletea model that shows a centered spinner
// while fn runs, then quits the instant it completes. It exists so blocking
// pre-fetch work (registry loads, manifest reads, adapter detection, ...)
// happens *inside* an alt-screen session instead of on the exposed terminal
// buffer between two tea.NewProgram lifecycles — that gap between programs,
// with real work running on the normal buffer during it, is what causes the
// visible "flash" when a dashboard command finishes and control returns to
// the launcher loop.
type loadingModel[T any] struct {
	theme *ui.Theme
	label string
	spin  spinner.Model
	fn    func() (T, error)

	width, height int
	result        T
	err           error
}

func newLoadingModel[T any](theme *ui.Theme, label string, fn func() (T, error)) loadingModel[T] {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = theme.Brand
	return loadingModel[T]{theme: theme, label: label, spin: sp, fn: fn}
}

func (m loadingModel[T]) Init() tea.Cmd {
	fn := m.fn
	return tea.Batch(m.spin.Tick, func() tea.Msg {
		res, err := fn()
		return loadingDoneMsg[T]{result: res, err: err}
	})
}

func (m loadingModel[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd
	case loadingDoneMsg[T]:
		m.result = msg.result
		m.err = msg.err
		return m, tea.Quit
	}
	return m, nil
}

func (m loadingModel[T]) View() string {
	th := m.theme
	inner := m.spin.View() + " " + th.Crumb.Render(m.label)
	if m.width <= 0 || m.height <= 0 {
		return inner
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, inner)
}

// RunWithLoading shows a centered spinner + label on its own alt-screen
// program while fn runs, then hands back fn's result the instant it's
// ready. Use this to wrap any blocking pre-fetch (registry load, manifest
// read, adapter detection, ...) that happens right before handing off to
// another alt-screen model — see the loadingModel doc comment for why this
// matters.
func RunWithLoading[T any](theme *ui.Theme, label string, fn func() (T, error)) (T, error) {
	out, err := Run(newLoadingModel(theme, label, fn))
	if err != nil {
		var zero T
		return zero, err
	}
	lm, ok := out.(loadingModel[T])
	if !ok {
		var zero T
		return zero, nil
	}
	return lm.result, lm.err
}

// RunWithLoadingIf is RunWithLoading gated by useTUI: when false (--json,
// --yes, piped, no TTY) there's no alt-screen to protect, so it just calls
// fn directly with no spinner. Callers that already computed
// tui.ShouldUseTUI for their own branching should reuse that result here
// instead of letting RunWithLoading attempt to open an alt-screen program
// without a TTY.
func RunWithLoadingIf[T any](useTUI bool, theme *ui.Theme, label string, fn func() (T, error)) (T, error) {
	if !useTUI {
		return fn()
	}
	return RunWithLoading(theme, label, fn)
}
