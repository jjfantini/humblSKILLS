package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

func newTestProgress(t *testing.T) ProgressModel {
	t.Helper()
	events := make(chan install.Event, 8)
	doneErr := make(chan error, 1)
	return NewProgressModel(ui.DefaultTheme(), "install", events, doneErr, 0)
}

func TestProgressModel_ApplyEventLifecycle(t *testing.T) {
	m := newTestProgress(t)

	// RunStart → total set.
	m.applyEvent(install.Event{Phase: install.PhaseRunStart, Total: 3})
	if m.total != 3 {
		t.Errorf("total = %d", m.total)
	}
	// TargetStart → current set, entry tracked.
	m.applyEvent(install.Event{
		Phase: install.PhaseTargetStart,
		Skill: "foo", Platform: "claude-code", Scope: "user",
	})
	if m.current == nil || m.current.skill != "foo" {
		t.Errorf("current = %+v", m.current)
	}
	// TargetDone → done incremented, outcome + path/version/store recorded.
	m.applyEvent(install.Event{
		Phase: install.PhaseTargetDone,
		Skill: "foo", Platform: "claude-code", Scope: "user",
		Outcome:   install.OutcomeInstalled,
		Path:      "/home/u/.claude/skills/foo",
		Version:   "1.2.3",
		StorePath: "/home/u/.humblskills/skills/foo",
	})
	if m.done != 1 {
		t.Errorf("done = %d", m.done)
	}
	if m.items[0].outcome != install.OutcomeInstalled {
		t.Errorf("outcome not recorded: %+v", m.items[0])
	}
	if m.items[0].path != "/home/u/.claude/skills/foo" {
		t.Errorf("path not recorded: %+v", m.items[0])
	}
	if m.items[0].version != "1.2.3" {
		t.Errorf("version not recorded: %+v", m.items[0])
	}
	if m.items[0].storePath != "/home/u/.humblskills/skills/foo" {
		t.Errorf("storePath not recorded: %+v", m.items[0])
	}
}

func TestProgressModel_RenderSummary_GroupsBySkillAndShowsStorePath(t *testing.T) {
	m := newTestProgress(t)
	m.width = 100
	m.running = false
	m.err = nil
	m.applyEvent(install.Event{Phase: install.PhaseRunStart, Total: 3})
	m.applyEvent(install.Event{
		Phase: install.PhaseTargetDone, Skill: "foo", Platform: "claude-code", Scope: "user",
		Outcome: install.OutcomeInstalled, Path: "/home/u/.claude/skills/foo",
		Version: "1.0.0", StorePath: "/home/u/.humblskills/skills/foo",
	})
	m.applyEvent(install.Event{
		Phase: install.PhaseTargetDone, Skill: "foo", Platform: "cursor", Scope: "user",
		Outcome: install.OutcomeInstalled, Path: "/home/u/.cursor/skills/foo",
		Version: "1.0.0", StorePath: "/home/u/.humblskills/skills/foo",
	})
	m.applyEvent(install.Event{
		Phase: install.PhaseTargetDone, Skill: "bar", Platform: "codex", Scope: "user",
		Outcome: install.OutcomeSkipped, Path: "/home/u/.agents/skills/bar",
		Version: "2.0.0", StorePath: "/home/u/.humblskills/skills/bar",
	})

	groups := m.groupedBySkill()
	if len(groups) != 2 {
		t.Fatalf("groups = %d, want 2: %+v", len(groups), groups)
	}
	if groups[0].skill != "foo" || len(groups[0].entries) != 2 {
		t.Errorf("foo group malformed: %+v", groups[0])
	}
	if groups[0].version != "1.0.0" || groups[0].storePath != "/home/u/.humblskills/skills/foo" {
		t.Errorf("foo group metadata wrong: %+v", groups[0])
	}
	if groups[1].skill != "bar" || len(groups[1].entries) != 1 {
		t.Errorf("bar group malformed: %+v", groups[1])
	}

	view := m.View()
	for _, want := range []string{
		"foo", "v1.0.0", "/home/u/.humblskills/skills/foo",
		"claude-code/user", "/home/u/.claude/skills/foo",
		"cursor/user", "/home/u/.cursor/skills/foo",
		"bar", "v2.0.0", "/home/u/.humblskills/skills/bar",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("summary view missing %q:\n%s", want, view)
		}
	}
}

func TestProgressModel_RenderSummary_EmptyWhenNothingInstalled(t *testing.T) {
	m := newTestProgress(t)
	m.width = 80
	m.running = false
	m.err = nil
	view := m.View()
	if !strings.Contains(view, "nothing to do") {
		t.Errorf("expected empty-summary message, got:\n%s", view)
	}
}

func TestProgressModel_DeduplicatesSameKey(t *testing.T) {
	m := newTestProgress(t)
	for i := 0; i < 3; i++ {
		m.applyEvent(install.Event{
			Phase: install.PhaseTargetStart,
			Skill: "foo", Platform: "x", Scope: "user",
		})
	}
	if len(m.items) != 1 {
		t.Errorf("items = %d, want 1", len(m.items))
	}
}

func TestProgressModel_ErrorPhaseCapturesErr(t *testing.T) {
	m := newTestProgress(t)
	boom := errors.New("boom")
	m.applyEvent(install.Event{
		Phase: install.PhaseError,
		Skill: "foo", Platform: "x", Scope: "user",
		Err: boom,
	})
	if m.items[0].errored != true {
		t.Error("expected errored=true")
	}
	if m.err == nil || m.err.Error() != "boom" {
		t.Errorf("err = %v", m.err)
	}
}

func TestProgressModel_UpsertReturnsExisting(t *testing.T) {
	m := newTestProgress(t)
	ev := install.Event{Skill: "foo", Platform: "x", Scope: "user"}
	a := m.upsert(ev)
	b := m.upsert(ev)
	if a != b {
		t.Error("upsert must return the same entry for the same key")
	}
}

func TestProgressModel_ViewMentionsCommandAndCounts(t *testing.T) {
	m := newTestProgress(t)
	m.width = 80
	m.applyEvent(install.Event{Phase: install.PhaseRunStart, Total: 2})
	m.applyEvent(install.Event{
		Phase: install.PhaseTargetStart,
		Skill: "foo", Platform: "x", Scope: "user",
	})
	v := m.View()
	if !strings.Contains(v, "install") {
		t.Errorf("view missing command header:\n%s", v)
	}
	if !strings.Contains(v, "0 / 2") {
		t.Errorf("view missing counter:\n%s", v)
	}
	if !strings.Contains(v, "foo") {
		t.Errorf("view missing skill:\n%s", v)
	}
}

func TestProgressModel_ViewClosedWithError(t *testing.T) {
	m := newTestProgress(t)
	m.width = 80
	m.err = errors.New("boom-err")
	m.running = false
	v := m.View()
	if !strings.Contains(v, "boom-err") {
		t.Errorf("view missing error:\n%s", v)
	}
	if !strings.Contains(v, "close") {
		t.Errorf("view should show close footer when done:\n%s", v)
	}
}

func TestSubscribe_ReadsEventThenDone(t *testing.T) {
	events := make(chan install.Event, 1)
	doneErr := make(chan error, 1)

	events <- install.Event{Phase: install.PhaseRunStart, Total: 1}
	cmd := Subscribe(events, doneErr)
	got := cmd()
	ev, ok := got.(ProgressEventMsg)
	if !ok {
		t.Fatalf("got %T, want ProgressEventMsg", got)
	}
	if ev.Event.Total != 1 {
		t.Errorf("event = %+v", ev.Event)
	}

	// Now close events and supply an error on doneErr.
	close(events)
	doneErr <- errors.New("final")
	got = Subscribe(events, doneErr)()
	done, ok := got.(ProgressDoneMsg)
	if !ok {
		t.Fatalf("got %T, want ProgressDoneMsg", got)
	}
	if done.Err == nil || done.Err.Error() != "final" {
		t.Errorf("err = %v", done.Err)
	}
}

func TestProgressModel_DoneMsg_DoesNotAutoQuit(t *testing.T) {
	// Regression: ProgressDoneMsg used to return tea.Quit unconditionally,
	// tearing the screen down the instant the engine finished — before the
	// user could read the result. It must now stay on screen until an
	// explicit keypress (see TestProgressModel_KeyMsgAfterDoneQuits).
	m := newTestProgress(t)
	m.running = true
	out, cmd := m.Update(ProgressDoneMsg{Err: nil})
	updated, ok := out.(ProgressModel)
	if !ok {
		t.Fatalf("Update returned %T, want ProgressModel", out)
	}
	if updated.running {
		t.Error("running should flip to false on ProgressDoneMsg")
	}
	if cmd != nil {
		t.Error("ProgressDoneMsg must not auto-quit — the screen should wait for a keypress")
	}
}

func TestProgressModel_DoneMsg_WithError_DoesNotAutoQuit(t *testing.T) {
	m := newTestProgress(t)
	m.running = true
	boom := errors.New("boom")
	out, cmd := m.Update(ProgressDoneMsg{Err: boom})
	updated := out.(ProgressModel)
	if updated.err == nil || updated.err.Error() != "boom" {
		t.Errorf("err = %v", updated.err)
	}
	if cmd != nil {
		t.Error("ProgressDoneMsg must not auto-quit even on error")
	}
}

func TestProgressModel_DoneMsg_Success_WithAutoReturn_StartsCountdown(t *testing.T) {
	events := make(chan install.Event, 8)
	doneErr := make(chan error, 1)
	m := NewProgressModel(ui.DefaultTheme(), "install", events, doneErr, 5*time.Second)
	m.running = true
	out, cmd := m.Update(ProgressDoneMsg{Err: nil})
	updated := out.(ProgressModel)
	if !updated.autoReturn.Active() {
		t.Error("expected the autoReturn timer to be armed on a successful done state")
	}
	if cmd == nil {
		t.Error("expected a tea.Cmd to schedule the first countdown tick")
	}
}

func TestProgressModel_DoneMsg_Error_WithAutoReturn_NeverStartsCountdown(t *testing.T) {
	events := make(chan install.Event, 8)
	doneErr := make(chan error, 1)
	m := NewProgressModel(ui.DefaultTheme(), "install", events, doneErr, 5*time.Second)
	m.running = true
	out, cmd := m.Update(ProgressDoneMsg{Err: errors.New("boom")})
	updated := out.(ProgressModel)
	if updated.autoReturn.Active() {
		t.Error("a failed run must never auto-return, regardless of the configured duration")
	}
	if cmd != nil {
		t.Error("error done state must not schedule an auto-quit")
	}
}

func TestProgressModel_AutoReturnTick_QuitsAfterDeadline(t *testing.T) {
	m := newTestProgress(t)
	m.running = false
	m.autoReturn = autoReturnTimer{duration: time.Millisecond, deadline: time.Now().Add(-time.Second)}
	_, cmd := m.Update(autoReturnTickMsg{})
	if cmd == nil {
		t.Fatal("expected a tea.Quit cmd once the deadline has passed")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("expected the elapsed countdown to return tea.Quit")
	}
}

func TestProgressModel_View_ShowsCountdownWhenAutoReturnActive(t *testing.T) {
	m := newTestProgress(t)
	m.running = false
	m.width = 80
	m.autoReturn = autoReturnTimer{duration: 5 * time.Second, deadline: time.Now().Add(5 * time.Second)}
	v := m.View()
	if !strings.Contains(v, "closing in") {
		t.Errorf("view should show the countdown when auto-return is active:\n%s", v)
	}
}

func TestProgressModel_KeyMsgAfterDoneQuits(t *testing.T) {
	m := newTestProgress(t)
	m.running = false
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("expected Quit cmd on enter after done")
	}
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Error("expected Quit cmd on q after done")
	}
}

func TestProgressEntry_Key(t *testing.T) {
	e := progressEntry{skill: "a", platform: "b", scope: "c"}
	if e.key() == "" {
		t.Error("key empty")
	}
	e2 := progressEntry{skill: "a", platform: "b", scope: "d"}
	if e.key() == e2.key() {
		t.Error("different scopes should differ")
	}
}
