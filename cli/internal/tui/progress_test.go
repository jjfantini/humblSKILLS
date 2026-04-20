package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

func newTestProgress(t *testing.T) ProgressModel {
	t.Helper()
	events := make(chan install.Event, 8)
	doneErr := make(chan error, 1)
	return NewProgressModel(ui.DefaultTheme(), "install", events, doneErr)
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
	// TargetDone → done incremented, outcome recorded.
	m.applyEvent(install.Event{
		Phase:    install.PhaseTargetDone,
		Skill:    "foo", Platform: "claude-code", Scope: "user",
		Outcome:  install.OutcomeInstalled,
	})
	if m.done != 1 {
		t.Errorf("done = %d", m.done)
	}
	if m.items[0].outcome != install.OutcomeInstalled {
		t.Errorf("outcome not recorded: %+v", m.items[0])
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
		Err:   boom,
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
