package tui

import (
	"os"
	"sync"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// This file implements the experimental single-program TUI "router". Normally
// every screen is its own tea.Program (its own alt-screen), so moving between
// panes tears the alternate-screen buffer down and back up — a visible flash to
// the shell. The router instead runs ONE long-lived program for the whole
// interactive session and swaps the active screen model in place, so the
// alt-screen is entered once and never torn down between panes.
//
// It's on by default for the interactive dashboard session (BeginSession),
// and off everywhere else: one-shot commands print results to the terminal
// after their screen exits, and a still-alive alt-screen would swallow those
// lines. Prompts (huh) are separate programs that can't share the host's
// terminal, so RunGuard releases/restores the host terminal around them (a
// flash at prompt time only, not navigation).

// RouterEnabled reports whether the single-program router is turned on. It
// always is, except when the HUMBLSKILLS_TUI_ROUTER env var is set to
// something other than "1" — an undocumented emergency escape hatch in case
// a terminal misbehaves with the long-lived alt-screen.
func RouterEnabled() bool {
	if v, ok := os.LookupEnv("HUMBLSKILLS_TUI_ROUTER"); ok {
		return v == "1"
	}
	return true
}

// sessionWanted is set once by BeginSession before any screen runs; Run only
// routes through the shared session program when it's true.
var sessionWanted bool

// BeginSession opts the current process into the single-program router (if
// enabled). Call it from long-lived interactive entry points — the dashboard
// loop — before the first screen. One-shot commands never call it, so they
// keep a per-screen program and their post-screen output stays visible.
func BeginSession() { sessionWanted = RouterEnabled() }

// --- messages driving the host --------------------------------------------

type activateMsg struct {
	model tea.Model
	done  chan tea.Model
}
type shutdownMsg struct{}
type childDoneMsg struct{ final tea.Model }

// sessionModel is the host: it renders and forwards messages to whichever
// screen model is currently active, and converts a screen's tea.Quit into a
// "child done" signal (returning that screen's final state to the caller)
// instead of quitting the whole program.
type sessionModel struct {
	active tea.Model
	size   tea.WindowSizeMsg
	done   chan tea.Model
}

func (m sessionModel) Init() tea.Cmd { return nil }

func (m sessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case activateMsg:
		m.active = msg.model
		m.done = msg.done
		cmd := m.active.Init()
		if m.size.Width > 0 { // give the new screen its size immediately
			var c2 tea.Cmd
			m.active, c2 = m.active.Update(m.size)
			cmd = tea.Batch(cmd, c2)
		}
		return m, cmd
	case childDoneMsg:
		if m.done != nil {
			m.done <- msg.final
			m.done = nil
		}
		m.active = msg.final // keep the last frame until the next screen activates
		return m, nil
	case shutdownMsg:
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.size = msg // record, then fall through to forward to the active screen
	}
	if m.active != nil {
		updated, cmd := m.active.Update(msg)
		m.active = updated
		return m, trapQuit(cmd, updated)
	}
	return m, nil
}

func (m sessionModel) View() string {
	if m.active != nil {
		return m.active.View()
	}
	return ""
}

// trapQuit wraps a child's command so a tea.Quit it emits becomes a childDoneMsg
// (ending just that screen) rather than a QuitMsg (which would end the whole
// session program). Only standalone quits are trapped — every screen model in
// this codebase quits standalone.
func trapQuit(cmd tea.Cmd, final tea.Model) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		msg := cmd()
		if _, ok := msg.(tea.QuitMsg); ok {
			return childDoneMsg{final: final}
		}
		return msg
	}
}

// --- session lifecycle -----------------------------------------------------

var (
	sessMu   sync.Mutex
	sessProg *tea.Program
	sessWG   sync.WaitGroup
)

func ensureSession() *tea.Program {
	sessMu.Lock()
	defer sessMu.Unlock()
	if sessProg != nil {
		return sessProg
	}
	ui.RunGuard = PauseForPrompt // route prompts through terminal release/restore
	p := tea.NewProgram(sessionModel{}, tea.WithAltScreen(), tea.WithMouseCellMotion())
	sessProg = p
	sessWG.Add(1)
	go func() {
		defer sessWG.Done()
		_, _ = p.Run()
		sessMu.Lock()
		sessProg = nil
		sessMu.Unlock()
	}()
	return p
}

// runOnSession activates a screen model on the shared session program and blocks
// until that screen finishes, returning its final model (mirrors Run's contract).
func runOnSession(m tea.Model) (tea.Model, error) {
	p := ensureSession()
	done := make(chan tea.Model, 1)
	p.Send(activateMsg{model: m, done: done})
	return <-done, nil
}

// Shutdown ends the interactive session program, restoring the terminal. Safe to
// call when no session is running (no-op). Call it once when the process is done
// with interactive work (wired via a defer around command execution).
func Shutdown() {
	sessMu.Lock()
	p := sessProg
	sessMu.Unlock()
	if p == nil {
		return
	}
	p.Send(shutdownMsg{})
	sessWG.Wait()
	ui.RunGuard = func(fn func() error) error { return fn() }
}

// PauseForPrompt releases the session's terminal so a separate prompt program
// (huh) can run, then restores it. No-op when no session is active.
func PauseForPrompt(fn func() error) error {
	sessMu.Lock()
	p := sessProg
	sessMu.Unlock()
	if p == nil {
		return fn()
	}
	p.ReleaseTerminal()
	defer p.RestoreTerminal()
	return fn()
}
