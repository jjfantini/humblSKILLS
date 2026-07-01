package tui

import (
	"testing"
	"time"
)

func TestAutoReturnTimer_ZeroDurationNeverStarts(t *testing.T) {
	var timer autoReturnTimer
	cmd := timer.Start()
	if cmd != nil {
		t.Error("Start with duration=0 must not return a tea.Cmd")
	}
	if timer.Active() {
		t.Error("timer with duration=0 must never be active")
	}
}

func TestAutoReturnTimer_StartArmsDeadlineAndReturnsCmd(t *testing.T) {
	timer := autoReturnTimer{duration: 5 * time.Second}
	cmd := timer.Start()
	if cmd == nil {
		t.Fatal("expected a tea.Cmd from Start")
	}
	if !timer.Active() {
		t.Error("expected timer to be active after Start")
	}
}

func TestAutoReturnTimer_TickBeforeDeadlineReschedules(t *testing.T) {
	timer := autoReturnTimer{duration: time.Hour}
	timer.Start()
	quit, cmd := timer.Tick()
	if quit {
		t.Error("must not quit before the deadline")
	}
	if cmd == nil {
		t.Error("expected the next tick to be scheduled")
	}
}

func TestAutoReturnTimer_TickAfterDeadlineQuits(t *testing.T) {
	timer := autoReturnTimer{duration: time.Millisecond}
	timer.deadline = time.Now().Add(-time.Second)
	quit, cmd := timer.Tick()
	if !quit {
		t.Error("expected quit=true once the deadline has passed")
	}
	if cmd != nil {
		t.Error("expected no further tick once quitting")
	}
}

func TestAutoReturnTimer_TickWhenInactiveIsNoop(t *testing.T) {
	var timer autoReturnTimer
	quit, cmd := timer.Tick()
	if quit || cmd != nil {
		t.Error("an inactive timer must never quit or schedule a tick")
	}
}

func TestAutoReturnTimer_RemainingSecondsRoundsUp(t *testing.T) {
	timer := autoReturnTimer{duration: 5 * time.Second}
	timer.deadline = time.Now().Add(2*time.Second + 100*time.Millisecond)
	if got := timer.RemainingSeconds(); got != 3 {
		t.Errorf("RemainingSeconds() = %d, want 3 (rounded up)", got)
	}
}

func TestAutoReturnTimer_RemainingSecondsZeroWhenInactiveOrElapsed(t *testing.T) {
	var inactive autoReturnTimer
	if got := inactive.RemainingSeconds(); got != 0 {
		t.Errorf("inactive timer RemainingSeconds() = %d, want 0", got)
	}

	elapsed := autoReturnTimer{duration: time.Second, deadline: time.Now().Add(-time.Second)}
	if got := elapsed.RemainingSeconds(); got != 0 {
		t.Errorf("elapsed timer RemainingSeconds() = %d, want 0", got)
	}
}
