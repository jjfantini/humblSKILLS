package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// ProgressEventMsg wraps an install.Event for the bubbletea runtime so the
// ProgressModel can react to engine progress without the caller exposing
// internal types.
type ProgressEventMsg struct{ Event install.Event }

// ProgressDoneMsg signals that the engine goroutine has finished. Err is nil
// on success; on failure it's the error the engine returned.
type ProgressDoneMsg struct{ Err error }

// Subscribe returns a tea.Cmd that reads the next engine event from ch. When
// the channel is closed, it emits ProgressDoneMsg{Err: doneErr} where doneErr
// is whatever the caller placed on the doneErr channel — typically the error
// returned by engine.Execute.
//
// Compose with bubbletea by re-issuing Subscribe from every ProgressEventMsg
// handler so the model keeps draining ch until it closes.
func Subscribe(ch <-chan install.Event, doneErr <-chan error) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			var err error
			select {
			case err = <-doneErr:
			default:
			}
			return ProgressDoneMsg{Err: err}
		}
		return ProgressEventMsg{Event: ev}
	}
}

// ProgressModel renders a framed progress bar + a scrollable per-target
// status list (or, once finished, the grouped summary) driven by install
// engine events. Use alongside a goroutine that wraps engine.Execute and
// forwards events onto a channel.
type ProgressModel struct {
	theme   *ui.Theme
	command string // breadcrumb detail, e.g. "install" or "update"
	bar     progress.Model
	spin    spinner.Model
	items   []*progressEntry
	keyed   map[string]*progressEntry
	total   int
	done    int
	width   int
	height  int
	current *progressEntry
	err     error
	running bool
	events  <-chan install.Event
	doneErr <-chan error

	// resultView owns the scrollable body (live target list, or the done
	// summary) plus the shared auto-return countdown — see scrollstatus.go.
	resultView scrollableDone
}

type progressEntry struct {
	skill    string
	platform string
	scope    string
	outcome  install.Outcome
	done     bool
	errored  bool
	// path is the platform-facing symlink target. version and storePath
	// (the canonical "source of truth" location) are filled in on
	// PhaseTargetDone — see applyEvent — and feed the grouped summary
	// rendered once the run finishes successfully.
	path      string
	version   string
	storePath string
}

func (p *progressEntry) key() string {
	return p.skill + "\x00" + p.platform + "\x00" + p.scope
}

// NewProgressModel builds a model subscribed to events/doneErr. command is
// the header breadcrumb ("install" or "update"). autoReturn is how long the
// done screen waits before auto-returning to the dashboard on success (0
// disables it - see profile.Profile.StatusAutoReturnDuration); a failed run
// always waits for the user regardless of this value, and even a successful
// run waits for the user to scroll to the bottom first if the summary
// doesn't already fit on screen — see scrollableDone.ArmIfReady.
func NewProgressModel(theme *ui.Theme, command string, events <-chan install.Event, doneErr <-chan error, autoReturn time.Duration) ProgressModel {
	p := progress.New(
		progress.WithGradient(string(theme.Palette.Brand), string(theme.Palette.Accent)),
		progress.WithoutPercentage(),
	)
	p.Width = 40

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = theme.Brand

	return ProgressModel{
		theme:      theme,
		command:    command,
		bar:        p,
		spin:       sp,
		keyed:      map[string]*progressEntry{},
		events:     events,
		doneErr:    doneErr,
		running:    true,
		resultView: newScrollableDone(autoReturn),
	}
}

// Err returns the engine's terminal error, if any. Meaningful only after the
// model has exited.
func (m ProgressModel) Err() error { return m.err }

func (m ProgressModel) Init() tea.Cmd {
	return tea.Batch(m.spin.Tick, Subscribe(m.events, m.doneErr))
}

func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		barW := m.width - 12
		if barW < 20 {
			barW = 20
		}
		if barW > 60 {
			barW = 60
		}
		m.bar.Width = barW
		m.refreshResultContent(false)
		return m, nil

	case ProgressEventMsg:
		m.applyEvent(msg.Event)
		cmds := []tea.Cmd{Subscribe(m.events, m.doneErr)}
		if m.total > 0 {
			cmds = append(cmds, m.bar.SetPercent(float64(m.done)/float64(m.total)))
		}
		m.refreshResultContent(false)
		return m, tea.Batch(cmds...)

	case ProgressDoneMsg:
		// Root cause of the "results flash by and vanish" bug: this used to
		// return tea.Quit unconditionally, so the screen tore itself down
		// the instant the engine finished — before the user could read
		// anything, success or failure. Now it just flips to the done
		// state (running=false) and stays on screen; the tea.KeyMsg case
		// below is what actually dismisses it, matching the "enter/q to
		// close" footer hint that was previously dead code. On success only,
		// ArmIfReady additionally arms a countdown that dismisses the screen
		// on its own once the summary has been fully seen — a failure
		// always waits for the user to read it.
		m.err = msg.Err
		m.running = false
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

	case progress.FrameMsg:
		newBar, cmd := m.bar.Update(msg)
		if nb, ok := newBar.(progress.Model); ok {
			m.bar = nb
		}
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

func (m ProgressModel) View() string {
	th := m.theme
	header := Header(th, HeaderSpec{
		Section: m.command,
		Meta:    m.resultView.ScrollIndicator(th),
	}, m.width)

	var top strings.Builder
	switch {
	case m.running:
		top.WriteString("  " + m.spin.View() + " " + th.Name.Render(fmt.Sprintf("%d / %d", m.done, m.total)) + "\n")
	case m.err != nil:
		top.WriteString("  " + th.Error.Render("✗ ") + th.Name.Render(fmt.Sprintf("%d / %d", m.done, m.total)) + "\n")
	default:
		top.WriteString("  " + th.Success.Render("✓ ") + th.Name.Render(fmt.Sprintf("%d / %d complete", m.done, m.total)) + "\n")
	}
	top.WriteString("  " + m.bar.View())

	var tail string
	if m.err != nil {
		tail = "\n  " + th.Error.Render("error: ") + th.Detail.Render(m.err.Error())
	}

	footer := m.renderFooter(th)

	return header + "\n\n" + top.String() + "\n\n" + m.resultView.View() + "\n" + tail + "\n" + footer
}

// renderFooter picks the key hints + right-anchored context for the current
// state: running (abort + scroll if the live list overflows), done with an
// error (close + scroll if there's more than the error line to read), or
// done successfully (close, plus whichever auto-return affordance applies).
func (m ProgressModel) renderFooter(th *ui.Theme) string {
	switch {
	case m.running:
		hints := []KeyHint{{Keys: "ctrl+c", Label: "abort"}}
		if m.resultView.Overflows() {
			hints = append(hints, KeyHint{Keys: "↑↓/pgup/pgdn", Label: "scroll"})
		}
		return Footer(th, hints, "", m.width)
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
// height minus the header, the fixed status+bar block above the body, the
// blank separator rows, the footer, and (when present) the error line below
// the body. Mirrors listdetail.go's bodyH computation.
func (m *ProgressModel) resizeResultView() {
	if m.width == 0 || m.height == 0 {
		return
	}
	const (
		headerH   = 2
		topH      = 2 // status line + bar
		footerH   = 2
		blankRows = 3 // header/top, top/body, body/footer separators
	)
	fixed := headerH + topH + footerH + blankRows
	if m.err != nil {
		fixed += 2 // blank + error line
	}
	bodyH := m.height - fixed
	if bodyH < 3 {
		bodyH = 3
	}
	m.resultView.Resize(m.width, bodyH)
}

// refreshResultContent recomputes the scrollable body's dimensions and
// content. resetToTop should be true exactly once — when the run reaches
// its terminal state — so the user reads the summary from the start; while
// running, the live list instead tails the bottom like a log.
func (m *ProgressModel) refreshResultContent(resetToTop bool) {
	m.resizeResultView()
	var body string
	if !m.running && m.err == nil {
		body = m.renderSummary()
	} else {
		body = m.renderItemsList()
	}
	m.resultView.SetContent(body, resetToTop)
	if m.running {
		m.resultView.FollowBottom()
	}
}

// renderItemsList renders one line per tracked target — used both while
// running (with the active entry highlighted) and, if the run ends in
// error, as the frozen final state of that same list.
func (m ProgressModel) renderItemsList() string {
	th := m.theme
	var sb strings.Builder
	for _, it := range m.items {
		active := m.running && m.current == it && !it.done && !it.errored
		line := m.renderEntry(it, active)
		prefix := "  "
		if active {
			prefix = th.Bullet.Render("▌") + " "
		}
		sb.WriteString(prefix + line + "\n")
	}
	return sb.String()
}

func (m *ProgressModel) applyEvent(ev install.Event) {
	switch ev.Phase {
	case install.PhaseRunStart:
		m.total = ev.Total
	case install.PhaseTargetStart:
		it := m.upsert(ev)
		m.current = it
	case install.PhaseTargetDone:
		it := m.upsert(ev)
		it.done = true
		it.outcome = ev.Outcome
		it.path = ev.Path
		it.version = ev.Version
		it.storePath = ev.StorePath
		m.done++
	case install.PhaseError:
		if ev.Skill != "" {
			it := m.upsert(ev)
			it.errored = true
		}
		if ev.Err != nil && m.err == nil {
			m.err = ev.Err
		}
	}
}

func (m *ProgressModel) upsert(ev install.Event) *progressEntry {
	it := &progressEntry{
		skill:    ev.Skill,
		platform: ev.Platform,
		scope:    ev.Scope,
	}
	k := it.key()
	if existing, ok := m.keyed[k]; ok {
		return existing
	}
	m.keyed[k] = it
	m.items = append(m.items, it)
	return it
}

func (m ProgressModel) renderEntry(it *progressEntry, active bool) string {
	th := m.theme
	var icon, label string
	switch {
	case it.errored:
		icon = th.DotNo.Render("●")
		label = th.Error.Render("error")
	case it.done:
		icon = th.DotOK.Render("●")
		label = th.Detail.Render(string(it.outcome))
	default:
		icon = th.DotWarn.Render("●")
		label = th.Detail.Render("running")
	}
	name := th.Name.Render(it.skill)
	if active {
		name = th.RowSelected.Render(it.skill)
	}
	where := ""
	if it.platform != "" {
		where = th.Platform.Render("[" + it.platform + "/" + it.scope + "]")
	}
	return fmt.Sprintf("%s %s %s %s", icon, name, where, label)
}

// skillSummaryGroup is every target this run touched for one skill, plus the
// version and canonical store path shared across them.
type skillSummaryGroup struct {
	skill     string
	version   string
	storePath string
	entries   []*progressEntry
}

// groupedBySkill buckets m.items by skill name, preserving first-seen order
// (the engine emits every target for one skill consecutively, so this is
// already a stable grouping, not just a sort).
func (m ProgressModel) groupedBySkill() []skillSummaryGroup {
	var groups []skillSummaryGroup
	idx := map[string]int{}
	for _, it := range m.items {
		i, ok := idx[it.skill]
		if !ok {
			i = len(groups)
			idx[it.skill] = i
			groups = append(groups, skillSummaryGroup{skill: it.skill})
		}
		g := &groups[i]
		if g.version == "" {
			g.version = it.version
		}
		if g.storePath == "" {
			g.storePath = it.storePath
		}
		g.entries = append(g.entries, it)
	}
	return groups
}

// renderSummary is the "what just happened" status screen shown once the run
// finishes successfully: one block per skill with its version, the
// canonical source-of-truth store location, and every platform symlink the
// run touched — the detail that used to only exist as stdout lines the
// dashboard loop immediately hid by re-entering the alt-screen.
func (m ProgressModel) renderSummary() string {
	th := m.theme
	groups := m.groupedBySkill()
	if len(groups) == 0 {
		return "  " + th.Detail.Render("nothing to do — every target was already up-to-date") + "\n"
	}
	var sb strings.Builder
	for i, g := range groups {
		if i > 0 {
			sb.WriteString("\n")
		}
		name := th.DetailTitle.Render(g.skill)
		ver := ""
		if g.version != "" {
			ver = "  " + th.DetailSub.Render("v"+g.version)
		}
		sb.WriteString("  " + name + ver + "\n")
		if g.storePath != "" {
			sb.WriteString("    " + th.KVKey.Render("installed to") + "  " + th.KVValue.Render(g.storePath) + "\n")
		}
		sb.WriteString("    " + th.SectionTitle.Render("SYMLINKED PLATFORMS") + "\n")
		for _, it := range g.entries {
			icon := th.DotOK.Render("●")
			label := th.Detail.Render(string(it.outcome))
			if it.errored {
				icon = th.DotNo.Render("●")
				label = th.Error.Render("error")
			}
			where := th.Platform.Render(it.platform + "/" + it.scope)
			sb.WriteString(fmt.Sprintf("      %s %s  %s  %s\n", icon, where, th.PathValue.Render(it.path), label))
		}
	}
	return sb.String()
}

// ExecuteWithProgress runs fn in a goroutine while a ProgressModel UI shows
// live engine progress. fn is expected to call engine.Execute (or equivalent)
// with an EventSink that forwards onto events. fn's error becomes
// ProgressDoneMsg.Err. autoReturn is how long the done screen waits before
// auto-returning to the dashboard on success (see NewProgressModel).
//
// Returns the final engine error when the model exits.
func ExecuteWithProgress(theme *ui.Theme, command string, autoReturn time.Duration, fn func(sink install.EventSink) error) error {
	events := make(chan install.Event, 32)
	doneErr := make(chan error, 1)

	sink := install.EventSink(func(ev install.Event) { events <- ev })

	go func() {
		defer close(events)
		doneErr <- fn(sink)
	}()

	m, err := Run(NewProgressModel(theme, command, events, doneErr, autoReturn))
	if err != nil {
		return err
	}
	if pm, ok := m.(ProgressModel); ok {
		return pm.Err()
	}
	return nil
}
