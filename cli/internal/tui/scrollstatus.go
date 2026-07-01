package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// scrollableDone renders a "what just happened" body of arbitrary length
// inside a viewport sized to whatever vertical space remains once the
// caller's fixed chrome (header, footer, any fixed lines above/below the
// body) is subtracted, and gates a shared autoReturnTimer so a completed
// screen never auto-dismisses before the user has actually seen everything:
// the countdown only arms once the body already fits on screen, or the user
// has scrolled all the way to the bottom. Embedded by value in ProgressModel
// and StatusModel so both share one implementation of "scroll to read it
// all, then auto-close (or don't, if you scrolled to read it)".
type scrollableDone struct {
	content viewport.Model
	auto    autoReturnTimer
	armed   bool
}

// newScrollableDone builds a zero-sized viewport (sized on the first
// Resize) paired with a countdown of the given duration (0 disables it,
// matching autoReturnTimer's own semantics).
func newScrollableDone(autoReturn time.Duration) scrollableDone {
	return scrollableDone{
		content: viewport.New(0, 0),
		auto:    autoReturnTimer{duration: autoReturn},
	}
}

// Resize updates the viewport dimensions. Call on tea.WindowSizeMsg and
// whenever the fixed chrome around the scroll region changes size (e.g. an
// error line appearing shrinks the body budget by one row).
func (s *scrollableDone) Resize(width, height int) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	s.content.Width = width
	s.content.Height = height
}

// SetContent replaces the body text. resetToTop should be true the moment a
// run transitions into its terminal state, so the user reads the summary
// from the beginning instead of wherever a live-updating list happened to
// leave the scroll position.
func (s *scrollableDone) SetContent(body string, resetToTop bool) {
	s.content.SetContent(body)
	if resetToTop {
		s.content.GotoTop()
	}
}

// FollowBottom keeps the viewport pinned to the newest content — used while
// a run is still in progress so the live per-target list behaves like a
// tailed log instead of freezing at whatever the scroll offset was when the
// list was shorter.
func (s *scrollableDone) FollowBottom() {
	s.content.GotoBottom()
}

// Overflows reports whether the body is taller than the viewport, i.e.
// whether there's anything to scroll at all.
func (s scrollableDone) Overflows() bool {
	return s.content.TotalLineCount() > s.content.Height
}

// ArmIfReady starts the countdown the first time the user has seen the
// entire body: either because it already fits without scrolling (AtBottom
// is trivially true the instant there's nothing to scroll), or because
// they've scrolled all the way down. ready gates this to the run's terminal
// success state — callers should pass false while running or on error, since
// a failed run must never auto-return. Once armed, it stays armed: scrolling
// back up afterward does not cancel the countdown, and repeated calls (e.g.
// holding a scroll key at the bottom) never re-extend the deadline.
func (s *scrollableDone) ArmIfReady(ready bool) tea.Cmd {
	if s.armed || !ready || !s.content.AtBottom() {
		return nil
	}
	s.armed = true
	return s.auto.Start()
}

// Tick advances the countdown; quit reports the screen should close now.
func (s *scrollableDone) Tick() (quit bool, cmd tea.Cmd) {
	return s.auto.Tick()
}

// Active reports whether the countdown is currently running (armed and not
// yet expired).
func (s scrollableDone) Active() bool { return s.auto.Active() }

// Enabled reports whether auto-return is configured at all (duration > 0),
// independent of whether it has armed yet — used to decide whether to show
// a "scroll to bottom to auto-close" hint versus no countdown affordance.
func (s scrollableDone) Enabled() bool { return s.auto.duration > 0 }

// RemainingSeconds is the whole seconds left before auto-quit.
func (s scrollableDone) RemainingSeconds() int { return s.auto.RemainingSeconds() }

// HandleKey applies a known scroll key to the viewport. ok reports whether
// key was a scroll key, so the caller can distinguish "consumed as a scroll"
// from "not a scroll key, handle it another way".
func (s *scrollableDone) HandleKey(key string) (ok bool) {
	switch key {
	case "up":
		s.content.LineUp(1)
	case "down":
		s.content.LineDown(1)
	case "pgup":
		s.content.ViewUp()
	case "pgdown":
		s.content.ViewDown()
	case "ctrl+u":
		s.content.HalfViewUp()
	case "ctrl+d":
		s.content.HalfViewDown()
	case "home":
		s.content.GotoTop()
	case "end":
		s.content.GotoBottom()
	default:
		return false
	}
	return true
}

// HandleMouse forwards a mouse event (wheel scroll) to the viewport.
func (s *scrollableDone) HandleMouse(msg tea.MouseMsg) tea.Cmd {
	var cmd tea.Cmd
	s.content, cmd = s.content.Update(msg)
	return cmd
}

// View renders the current viewport window.
func (s scrollableDone) View() string { return s.content.View() }

// ScrollIndicator returns a compact "▲▼ NN%" widget when the body overflows
// the viewport, or "" when everything is already visible — mirrors
// listdetail.go's scrollIndicator so both screens share the same affordance.
func (s scrollableDone) ScrollIndicator(th *ui.Theme) string {
	if !s.Overflows() {
		return ""
	}
	up, down := "△", "▽"
	if !s.content.AtTop() {
		up = "▲"
	}
	if !s.content.AtBottom() {
		down = "▼"
	}
	pct := int(s.content.ScrollPercent()*100 + 0.5)
	return th.Meta.Render(fmt.Sprintf("%s%s %d%%", up, down, pct))
}
