package tui

import (
	"fmt"
	"strings"

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

// ProgressModel renders a framed progress bar + per-target status list driven
// by install engine events. Use alongside a goroutine that wraps
// engine.Execute and forwards events onto a channel.
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
	current *progressEntry
	err     error
	running bool
	events  <-chan install.Event
	doneErr <-chan error
}

type progressEntry struct {
	skill    string
	platform string
	scope    string
	outcome  install.Outcome
	done     bool
	errored  bool
}

func (p *progressEntry) key() string {
	return p.skill + "\x00" + p.platform + "\x00" + p.scope
}

// NewProgressModel builds a model subscribed to events/doneErr. command is the
// header breadcrumb ("install" or "update").
func NewProgressModel(theme *ui.Theme, command string, events <-chan install.Event, doneErr <-chan error) ProgressModel {
	p := progress.New(
		progress.WithGradient(string(theme.Palette.Brand), string(theme.Palette.Accent)),
		progress.WithoutPercentage(),
	)
	p.Width = 40

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = theme.Brand

	return ProgressModel{
		theme:   theme,
		command: command,
		bar:     p,
		spin:    sp,
		keyed:   map[string]*progressEntry{},
		events:  events,
		doneErr: doneErr,
		running: true,
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
		barW := m.width - 12
		if barW < 20 {
			barW = 20
		}
		if barW > 60 {
			barW = 60
		}
		m.bar.Width = barW
		return m, nil

	case ProgressEventMsg:
		m.applyEvent(msg.Event)
		cmds := []tea.Cmd{Subscribe(m.events, m.doneErr)}
		if m.total > 0 {
			cmds = append(cmds, m.bar.SetPercent(float64(m.done)/float64(m.total)))
		}
		return m, tea.Batch(cmds...)

	case ProgressDoneMsg:
		m.err = msg.Err
		m.running = false
		return m, tea.Quit

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

	case tea.KeyMsg:
		if !m.running && (msg.String() == "q" || msg.String() == "enter" || msg.String() == "esc") {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ProgressModel) View() string {
	th := m.theme
	header := Header(th, HeaderSpec{Section: m.command}, m.width)

	var sb strings.Builder
	if m.running {
		sb.WriteString("  " + m.spin.View() + " " +
			th.Name.Render(fmt.Sprintf("%d / %d", m.done, m.total)) + "\n")
	} else if m.err != nil {
		sb.WriteString("  " + th.Error.Render("✗ ") +
			th.Name.Render(fmt.Sprintf("%d / %d", m.done, m.total)) + "\n")
	} else {
		sb.WriteString("  " + th.Success.Render("✓ ") +
			th.Name.Render(fmt.Sprintf("%d / %d complete", m.done, m.total)) + "\n")
	}
	sb.WriteString("  " + m.bar.View() + "\n\n")

	for _, it := range m.items {
		active := m.running && m.current == it && !it.done && !it.errored
		line := m.renderEntry(it, active)
		prefix := "  "
		if active {
			prefix = th.Bullet.Render("▌") + " "
		}
		sb.WriteString(prefix + line + "\n")
	}

	if m.err != nil {
		sb.WriteString("\n  " + th.Error.Render("error: ") + th.Detail.Render(m.err.Error()) + "\n")
	}

	var footer string
	if m.running {
		footer = Footer(th, []KeyHint{{Keys: "ctrl+c", Label: "abort"}}, "", m.width)
	} else {
		footer = Footer(th, []KeyHint{{Keys: "enter/q", Label: "close"}}, "", m.width)
	}
	return header + "\n\n" + sb.String() + "\n" + footer
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

// ExecuteWithProgress runs fn in a goroutine while a ProgressModel UI shows
// live engine progress. fn is expected to call engine.Execute (or equivalent)
// with an EventSink that forwards onto events. fn's error becomes
// ProgressDoneMsg.Err.
//
// Returns the final engine error when the model exits.
func ExecuteWithProgress(theme *ui.Theme, command string, fn func(sink install.EventSink) error) error {
	events := make(chan install.Event, 32)
	doneErr := make(chan error, 1)

	sink := install.EventSink(func(ev install.Event) { events <- ev })

	go func() {
		defer close(events)
		doneErr <- fn(sink)
	}()

	m, err := Run(NewProgressModel(theme, command, events, doneErr))
	if err != nil {
		return err
	}
	if pm, ok := m.(ProgressModel); ok {
		return pm.Err()
	}
	return nil
}

