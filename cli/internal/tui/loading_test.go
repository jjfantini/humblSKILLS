package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

func TestLoadingModel_DoneMsg_CarriesResultAndQuits(t *testing.T) {
	m := newLoadingModel(ui.DefaultTheme(), "loading…", func() (int, error) { return 42, nil })
	out, cmd := m.Update(loadingDoneMsg[int]{result: 42, err: nil})
	updated, ok := out.(loadingModel[int])
	if !ok {
		t.Fatalf("Update returned %T, want loadingModel[int]", out)
	}
	if updated.result != 42 {
		t.Errorf("result = %d, want 42", updated.result)
	}
	if cmd == nil {
		t.Fatal("expected a tea.Quit cmd once loading finishes")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("expected loadingDoneMsg to trigger tea.Quit")
	}
}

func TestLoadingModel_DoneMsg_CarriesError(t *testing.T) {
	m := newLoadingModel(ui.DefaultTheme(), "loading…", func() (int, error) { return 0, nil })
	boom := errors.New("boom")
	out, _ := m.Update(loadingDoneMsg[int]{err: boom})
	updated := out.(loadingModel[int])
	if updated.err == nil || updated.err.Error() != "boom" {
		t.Errorf("err = %v", updated.err)
	}
}

func TestLoadingModel_View_ShowsLabel(t *testing.T) {
	m := newLoadingModel(ui.DefaultTheme(), "scanning environment…", func() (int, error) { return 0, nil })
	m.width, m.height = 80, 24
	v := m.View()
	if !strings.Contains(v, "scanning environment…") {
		t.Errorf("view missing label:\n%s", v)
	}
}

func TestRunWithLoadingIf_FalseCallsFnDirectlyWithNoProgram(t *testing.T) {
	called := false
	got, err := RunWithLoadingIf(false, ui.DefaultTheme(), "loading…", func() (int, error) {
		called = true
		return 7, nil
	})
	if err != nil {
		t.Fatalf("RunWithLoadingIf(false, ...): %v", err)
	}
	if !called {
		t.Error("expected fn to be called directly when useTUI=false")
	}
	if got != 7 {
		t.Errorf("got %d, want 7", got)
	}
}

func TestRunWithLoadingIf_FalsePropagatesError(t *testing.T) {
	boom := errors.New("boom")
	_, err := RunWithLoadingIf(false, ui.DefaultTheme(), "loading…", func() (int, error) {
		return 0, boom
	})
	if err != boom {
		t.Errorf("err = %v, want %v", err, boom)
	}
}

func TestLoadingModel_Init_RunsFnAndReturnsDoneMsg(t *testing.T) {
	m := newLoadingModel(ui.DefaultTheme(), "loading…", func() (string, error) { return "hello", nil })
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected Init to return a batched cmd")
	}
	// Init batches spin.Tick with the fn-running cmd; run the batch and look
	// for the loadingDoneMsg among the resulting messages.
	msg := cmd()
	found := false
	var walk func(m tea.Msg)
	walk = func(m tea.Msg) {
		switch mm := m.(type) {
		case tea.BatchMsg:
			for _, c := range mm {
				walk(c())
			}
		case loadingDoneMsg[string]:
			if mm.result != "hello" {
				t.Errorf("result = %q, want hello", mm.result)
			}
			found = true
		}
	}
	walk(msg)
	if !found {
		t.Error("expected a loadingDoneMsg[string] somewhere in Init's batched commands")
	}
}
