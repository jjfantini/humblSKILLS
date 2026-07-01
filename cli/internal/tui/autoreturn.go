package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// autoReturnTickMsg drives the 1-second countdown shared by every "stays on
// screen until dismissed, or auto-returns after Ns" done screen
// (ProgressModel, StatusModel).
type autoReturnTickMsg struct{}

// autoReturnTimer tracks a single countdown-to-auto-quit deadline. Embed by
// value in any model that wants "wait for a keypress, or auto-dismiss after
// N seconds" behavior once it reaches a terminal success state. A zero-value
// autoReturnTimer never auto-quits — callers must set duration (e.g. via
// profile.Profile.StatusAutoReturnDuration) before calling Start.
type autoReturnTimer struct {
	duration time.Duration // 0 = disabled, never auto-quits
	deadline time.Time     // zero until Start is called
}

// Start arms the timer (if duration > 0) and returns the tea.Cmd that kicks
// off the first tick. Call once, when the model reaches its terminal success
// state — never on failure, since the user needs to actually read the error.
func (t *autoReturnTimer) Start() tea.Cmd {
	if t.duration <= 0 {
		return nil
	}
	t.deadline = time.Now().Add(t.duration)
	return autoReturnTick()
}

// Active reports whether a countdown is currently running.
func (t autoReturnTimer) Active() bool {
	return t.duration > 0 && !t.deadline.IsZero()
}

// Tick advances the countdown on every autoReturnTickMsg. Returns quit=true
// once the deadline has passed (the caller should return tea.Quit);
// otherwise it returns the tea.Cmd to schedule the next tick.
func (t autoReturnTimer) Tick() (quit bool, cmd tea.Cmd) {
	if !t.Active() {
		return false, nil
	}
	if !time.Now().Before(t.deadline) {
		return true, nil
	}
	return false, autoReturnTick()
}

// RemainingSeconds returns the whole seconds left before auto-quit, rounded
// up so the countdown never displays "0s" before it actually fires.
func (t autoReturnTimer) RemainingSeconds() int {
	if !t.Active() {
		return 0
	}
	remaining := time.Until(t.deadline)
	if remaining <= 0 {
		return 0
	}
	secs := int(remaining / time.Second)
	if remaining%time.Second > 0 {
		secs++
	}
	return secs
}

func autoReturnTick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return autoReturnTickMsg{} })
}
