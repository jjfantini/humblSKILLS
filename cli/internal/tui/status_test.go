package tui

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

func newTestStatus(autoReturn time.Duration) StatusModel {
	return NewStatusModel(ui.DefaultTheme(), "registry", "refreshing registry…", autoReturn, func() (StatusResult, error) {
		return StatusResult{}, nil
	})
}

func TestStatusModel_DoneMsg_Success_NoAutoReturn_DoesNotAutoQuit(t *testing.T) {
	m := newTestStatus(0)
	m.running = true
	out, cmd := m.Update(statusDoneMsg{result: StatusResult{Headline: "ok"}})
	updated, ok := out.(StatusModel)
	if !ok {
		t.Fatalf("Update returned %T, want StatusModel", out)
	}
	if updated.running {
		t.Error("running should flip to false on statusDoneMsg")
	}
	if cmd != nil {
		t.Error("with autoReturn=0 the screen must wait for a keypress, not auto-quit")
	}
}

func TestStatusModel_DoneMsg_Success_WithAutoReturn_StartsCountdown(t *testing.T) {
	m := newTestStatus(5 * time.Second)
	m.running = true
	// An empty result fits without scrolling, so the countdown arms
	// immediately — matching today's behavior for short results.
	m.width, m.height = 80, 40
	out, cmd := m.Update(statusDoneMsg{result: StatusResult{Headline: "ok"}})
	updated := out.(StatusModel)
	if !updated.resultView.Active() {
		t.Error("expected the auto-return timer to be armed on a successful result that already fits on screen")
	}
	if cmd == nil {
		t.Error("expected a tea.Cmd to schedule the first tick")
	}
}

// TestStatusModel_ArmIfReady_WaitsForScrollWhenContentOverflows mirrors the
// equivalent ProgressModel test: a result taller than the viewport must not
// auto-return until the user has scrolled all the way down to see it.
func TestStatusModel_ArmIfReady_WaitsForScrollWhenContentOverflows(t *testing.T) {
	m := newTestStatus(5 * time.Second)
	m.running = true
	m.width, m.height = 80, 12

	lines := make([]string, 30)
	for i := range lines {
		lines[i] = fmt.Sprintf("detail line %d", i)
	}

	out, cmd := m.Update(statusDoneMsg{result: StatusResult{Headline: "ok", Lines: lines}})
	updated := out.(StatusModel)
	if !updated.resultView.Overflows() {
		t.Fatalf("expected 30 detail lines to overflow a %d-row viewport", updated.resultView.content.Height)
	}
	if updated.resultView.Active() || cmd != nil {
		t.Fatal("auto-return must not arm before the user has scrolled to see the whole result")
	}

	out, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyEnd})
	updated = out.(StatusModel)
	if !updated.resultView.Active() {
		t.Error("expected auto-return to arm once the user scrolled to the bottom")
	}
	if cmd == nil {
		t.Error("expected a tea.Cmd to schedule the first countdown tick after reaching the bottom")
	}
}

func TestStatusModel_DoneMsg_Error_NeverAutoReturns(t *testing.T) {
	m := newTestStatus(5 * time.Second)
	m.running = true
	boom := errors.New("boom")
	out, cmd := m.Update(statusDoneMsg{err: boom})
	updated := out.(StatusModel)
	if updated.resultView.Active() {
		t.Error("a failed run must never auto-return — the user needs to read the error")
	}
	if cmd != nil {
		t.Error("error done state must not schedule an auto-quit")
	}
	if updated.err == nil || updated.err.Error() != "boom" {
		t.Errorf("err = %v", updated.err)
	}
}

func TestStatusModel_AutoReturnTick_QuitsAfterDeadline(t *testing.T) {
	m := newTestStatus(1 * time.Millisecond)
	m.running = false
	m.resultView.auto.deadline = time.Now().Add(-time.Second) // already elapsed
	_, cmd := m.Update(autoReturnTickMsg{})
	if cmd == nil {
		t.Fatal("expected tea.Quit cmd once the deadline has passed")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("expected the elapsed countdown to return tea.Quit")
	}
}

func TestStatusModel_KeyMsgAfterDoneQuits(t *testing.T) {
	m := newTestStatus(0)
	m.running = false
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("expected Quit cmd on enter after done")
	}
}

func TestStatusModel_View_ShowsHeadlineAndDetailLines(t *testing.T) {
	m := newTestStatus(0)
	m.running = false
	m.result = StatusResult{Headline: "registry refreshed: 11 skills", Lines: []string{"cache: /tmp/registry.json"}}
	m.width, m.height = 80, 40
	m.refreshResultContent(true)
	v := m.View()
	if !strings.Contains(v, "registry refreshed: 11 skills") {
		t.Errorf("view missing headline:\n%s", v)
	}
	if !strings.Contains(v, "cache: /tmp/registry.json") {
		t.Errorf("view missing detail line:\n%s", v)
	}
	if !strings.Contains(v, "close") {
		t.Errorf("view should show close footer when done:\n%s", v)
	}
}

func TestStatusModel_View_ShowsCountdownWhenAutoReturnActive(t *testing.T) {
	m := newTestStatus(5 * time.Second)
	m.running = false
	m.resultView.auto.deadline = time.Now().Add(5 * time.Second)
	m.width, m.height = 80, 40
	v := m.View()
	if !strings.Contains(v, "closing in") {
		t.Errorf("view should show the countdown when auto-return is active:\n%s", v)
	}
}

func TestStatusModel_View_ShowsError(t *testing.T) {
	m := newTestStatus(0)
	m.running = false
	m.err = errors.New("registry unreachable")
	m.width, m.height = 80, 40
	m.refreshResultContent(true)
	v := m.View()
	if !strings.Contains(v, "registry unreachable") {
		t.Errorf("view missing error:\n%s", v)
	}
}
