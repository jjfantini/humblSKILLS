package tui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/maaslalani/confetty/confetti"
	"github.com/maaslalani/confetty/fireworks"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// UpgradeStep names one ordered step in the `upgrade` command's own
// pipeline. Distinct from internal/install's Phase — this models the CLI
// upgrading *itself*, not a skill install/update run.
type UpgradeStep string

const (
	UpgradeStepCheckingLatest    UpgradeStep = "checking_latest"
	UpgradeStepBrewUpdating      UpgradeStep = "brew_updating"
	UpgradeStepBrewUpgrading     UpgradeStep = "brew_upgrading"
	UpgradeStepDownloading       UpgradeStep = "downloading"
	UpgradeStepVerifyingChecksum UpgradeStep = "verifying_checksum"
	UpgradeStepInstalling        UpgradeStep = "installing"
	UpgradeStepVerifyingInstall  UpgradeStep = "verifying_install"
)

// upgradeStepLabels is the display label for every step that could appear.
// A given run only walks a subset — the Homebrew path skips
// download/verify/install in favour of UpgradeStepBrewUpdating +
// UpgradeStepBrewUpgrading.
var upgradeStepLabels = map[UpgradeStep]string{
	UpgradeStepCheckingLatest:    "checking latest release",
	UpgradeStepBrewUpdating:      "running brew update",
	UpgradeStepBrewUpgrading:     "running brew upgrade humblskills",
	UpgradeStepDownloading:       "downloading release archive",
	UpgradeStepVerifyingChecksum: "verifying checksum",
	UpgradeStepInstalling:        "installing",
	UpgradeStepVerifyingInstall:  "verifying installed version",
}

// UpgradeEvent is one progress notification the upgrade pipeline reports.
// An event with Err == nil means "entering Step"; the previously-running
// step (if any) is implicitly marked done.
type UpgradeEvent struct {
	Step UpgradeStep
	Err  error
}

// UpgradeDoneMsg signals the goroutine driving the upgrade finished. Err is
// nil on success.
type UpgradeDoneMsg struct{ Err error }

// upgradeSubscribe returns a tea.Cmd that reads the next event from ch.
func upgradeSubscribe(ch <-chan UpgradeEvent, doneErr <-chan error) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			var err error
			select {
			case err = <-doneErr:
			default:
			}
			return UpgradeDoneMsg{Err: err}
		}
		return ev
	}
}

type upgradeStepStatus int

const (
	upgradeStepPending upgradeStepStatus = iota
	upgradeStepRunning
	upgradeStepDone
	upgradeStepFailed
)

type upgradeStepEntry struct {
	step   UpgradeStep
	status upgradeStepStatus
}

// celebrationDuration is how long the confetti/fireworks animation plays
// after a successful upgrade before the program quits on its own.
const celebrationDuration = 2500 * time.Millisecond

type celebrationTimeoutMsg struct{}

// newCelebration picks confetti or fireworks at random for a little
// variety — either way it's a few seconds of fun before the program exits.
func newCelebration() tea.Model {
	if rand.Intn(2) == 0 {
		return confetti.InitialModel()
	}
	return fireworks.InitialModel()
}

// UpgradeModel renders the `upgrade` command's themed step list (current
// version -> latest version, live progress, success/failure) and, on a
// successful upgrade, hands off to a confetti/fireworks celebration before
// exiting on its own.
type UpgradeModel struct {
	theme *ui.Theme

	fromVersion string
	toVersion   string

	steps []*upgradeStepEntry
	spin  spinner.Model

	width, height int

	done bool
	err  error

	celebrating bool
	celebration tea.Model

	events  <-chan UpgradeEvent
	doneErr <-chan error
}

// NewUpgradeModel builds a model subscribed to events/doneErr. plannedSteps
// is the ordered subset of steps this particular run will walk (Homebrew vs
// self-download paths differ).
func NewUpgradeModel(theme *ui.Theme, fromVersion, toVersion string, plannedSteps []UpgradeStep, events <-chan UpgradeEvent, doneErr <-chan error) UpgradeModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = theme.Brand

	steps := make([]*upgradeStepEntry, 0, len(plannedSteps))
	for _, s := range plannedSteps {
		steps = append(steps, &upgradeStepEntry{step: s})
	}

	return UpgradeModel{
		theme:       theme,
		fromVersion: fromVersion,
		toVersion:   toVersion,
		steps:       steps,
		spin:        sp,
		events:      events,
		doneErr:     doneErr,
	}
}

// Err returns the pipeline's terminal error, if any. Meaningful only after
// the model has exited.
func (m UpgradeModel) Err() error { return m.err }

func (m UpgradeModel) Init() tea.Cmd {
	return tea.Batch(m.spin.Tick, upgradeSubscribe(m.events, m.doneErr))
}

func (m UpgradeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.celebrating {
			var cmd tea.Cmd
			m.celebration, cmd = m.celebration.Update(msg)
			return m, cmd
		}
		return m, nil

	case UpgradeEvent:
		m.applyEvent(msg)
		return m, upgradeSubscribe(m.events, m.doneErr)

	case UpgradeDoneMsg:
		m.done = true
		m.err = msg.Err
		for _, s := range m.steps {
			if s.status != upgradeStepRunning {
				continue
			}
			if m.err != nil {
				s.status = upgradeStepFailed
			} else {
				s.status = upgradeStepDone
			}
		}
		if m.err != nil {
			return m, nil
		}
		m.celebrating = true
		m.celebration = newCelebration()
		width, height := m.width, m.height
		sizeCmd := func() tea.Msg { return tea.WindowSizeMsg{Width: width, Height: height} }
		return m, tea.Batch(m.celebration.Init(), sizeCmd, celebrationTimeout())

	case celebrationTimeoutMsg:
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q", "enter", "esc":
			if m.done {
				return m, tea.Quit
			}
		}
		if m.celebrating {
			var cmd tea.Cmd
			m.celebration, cmd = m.celebration.Update(msg)
			return m, cmd
		}
		return m, nil

	default:
		if m.celebrating {
			var cmd tea.Cmd
			m.celebration, cmd = m.celebration.Update(msg)
			return m, cmd
		}
		return m, nil
	}
}

// versionTag renders a version for display: "v2.17.0" for ordinary semver
// builds, but the bare string itself for non-numeric local builds ("dev")
// so the UI doesn't show the slightly silly "vdev".
func versionTag(v string) string {
	if v != "" && v[0] >= '0' && v[0] <= '9' {
		return "v" + v
	}
	return v
}

func celebrationTimeout() tea.Cmd {
	return tea.Tick(celebrationDuration, func(time.Time) tea.Msg { return celebrationTimeoutMsg{} })
}

func (m *UpgradeModel) applyEvent(ev UpgradeEvent) {
	// A new event always means we've moved past whatever was running.
	for _, s := range m.steps {
		if s.status == upgradeStepRunning {
			s.status = upgradeStepDone
		}
	}
	for _, s := range m.steps {
		if s.step != ev.Step {
			continue
		}
		if ev.Err != nil {
			s.status = upgradeStepFailed
		} else {
			s.status = upgradeStepRunning
		}
	}
}

func (m UpgradeModel) View() string {
	th := m.theme

	if m.celebrating {
		msg := "  " + th.Success.Render("✓ ") + th.Name.Render(fmt.Sprintf("humblskills is now %s", versionTag(m.toVersion)))
		return msg + "\n\n" + m.celebration.View()
	}

	header := Header(th, HeaderSpec{Section: "Upgrade"}, m.width)

	var sb strings.Builder
	switch {
	case m.done && m.err != nil:
		sb.WriteString("  " + th.Error.Render("✗ ") + th.Name.Render("upgrade failed") + "\n\n")
	case m.done:
		sb.WriteString("  " + th.Success.Render("✓ ") + th.Name.Render(versionTag(m.toVersion)) + "\n\n")
	default:
		sb.WriteString("  " + m.spin.View() + " " +
			th.Name.Render(versionTag(m.fromVersion)) + th.Detail.Render(" → ") + th.Name.Render(versionTag(m.toVersion)) + "\n\n")
	}

	for _, s := range m.steps {
		sb.WriteString("  " + renderUpgradeStep(th, s) + "\n")
	}

	if m.err != nil {
		sb.WriteString("\n  " + th.Error.Render("error: ") + th.Detail.Render(m.err.Error()) + "\n")
	}

	footer := Footer(th, []KeyHint{{Keys: "ctrl+c", Label: "abort"}}, "", m.width)
	if m.done {
		footer = Footer(th, []KeyHint{{Keys: "enter/q", Label: "close"}}, "", m.width)
	}
	return header + "\n\n" + sb.String() + "\n" + footer
}

func renderUpgradeStep(th *ui.Theme, e *upgradeStepEntry) string {
	label := upgradeStepLabels[e.step]
	switch e.status {
	case upgradeStepDone:
		return th.DotOK.Render("●") + " " + th.Detail.Render(label)
	case upgradeStepFailed:
		return th.DotNo.Render("●") + " " + th.Error.Render(label)
	case upgradeStepRunning:
		return th.DotWarn.Render("●") + " " + th.Name.Render(label)
	default:
		return th.RowDim.Render("○ " + label)
	}
}

// ExecuteUpgrade runs fn in a goroutine while an UpgradeModel shows live
// progress through plannedSteps, the version transition, and — on success —
// a confetti/fireworks celebration. fn is expected to call into
// internal/selfupdate and report progress through sink. Returns the
// pipeline's terminal error (nil on success).
func ExecuteUpgrade(theme *ui.Theme, fromVersion, toVersion string, plannedSteps []UpgradeStep, fn func(sink func(UpgradeEvent)) error) error {
	events := make(chan UpgradeEvent, 16)
	doneErr := make(chan error, 1)

	go func() {
		defer close(events)
		doneErr <- fn(func(ev UpgradeEvent) { events <- ev })
	}()

	m, err := Run(NewUpgradeModel(theme, fromVersion, toVersion, plannedSteps, events, doneErr))
	if err != nil {
		return err
	}
	if um, ok := m.(UpgradeModel); ok {
		return um.Err()
	}
	return nil
}
