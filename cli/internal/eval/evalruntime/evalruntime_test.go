package evalruntime_test

import (
	"context"
	"sort"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/evalruntime"
	"github.com/jjfantini/humblSKILLS/cli/internal/secrets"
	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

// DefaultRegistry is a small wiring function. These tests protect the
// contract callers rely on: every runner lines up in a stable order,
// every name is unique, and lookups / auto-pick fail cleanly when no
// runner is available (common on headless CI).
func TestDefaultRegistry_ReturnsSixRunnersInStableOrder(t *testing.T) {
	s := testutil.NewSandbox(t)
	testutil.UseFakeKeyring(t)

	store, err := secrets.NewStore(s.SecretsPath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	reg := evalruntime.DefaultRegistry(store)
	if reg == nil {
		t.Fatal("DefaultRegistry returned nil")
	}

	runners := reg.All()
	if len(runners) != 6 {
		t.Fatalf("got %d runners, want 6", len(runners))
	}

	// Order matters: CLI runners first (lowest friction), then API
	// runners, then mock. AutoPick walks this list in order.
	wantOrder := []string{"claudecode", "cursor-agent", "codex", "anthropic-api", "openai-api", "mock"}
	for i, want := range wantOrder {
		if got := runners[i].Name(); got != want {
			t.Errorf("position %d: got %q, want %q", i, got, want)
		}
	}
}

func TestDefaultRegistry_NamesAreUnique(t *testing.T) {
	s := testutil.NewSandbox(t)
	testutil.UseFakeKeyring(t)
	store, _ := secrets.NewStore(s.SecretsPath)

	reg := evalruntime.DefaultRegistry(store)
	seen := map[string]bool{}
	for _, r := range reg.All() {
		if seen[r.Name()] {
			t.Errorf("duplicate runner name: %q", r.Name())
		}
		seen[r.Name()] = true
	}
}

func TestDefaultRegistry_NamesReturnsSorted(t *testing.T) {
	s := testutil.NewSandbox(t)
	testutil.UseFakeKeyring(t)
	store, _ := secrets.NewStore(s.SecretsPath)

	reg := evalruntime.DefaultRegistry(store)
	names := reg.Names()
	sorted := append([]string(nil), names...)
	sort.Strings(sorted)
	for i := range names {
		if names[i] != sorted[i] {
			t.Errorf("Names() not sorted: %v", names)
			break
		}
	}
}

func TestDefaultRegistry_ByNameLookup(t *testing.T) {
	s := testutil.NewSandbox(t)
	testutil.UseFakeKeyring(t)
	store, _ := secrets.NewStore(s.SecretsPath)

	reg := evalruntime.DefaultRegistry(store)

	// The mock runner is always available and has no external deps,
	// so it's the safe target for a positive lookup test.
	r, err := reg.ByName("mock")
	if err != nil {
		t.Fatalf("ByName(mock): %v", err)
	}
	if r.Name() != "mock" {
		t.Errorf("ByName returned %q", r.Name())
	}

	if _, err := reg.ByName("nonexistent"); err == nil {
		t.Error("ByName(nonexistent) should error")
	}
}

func TestDefaultRegistry_MockRunnerIsAlwaysAvailable(t *testing.T) {
	s := testutil.NewSandbox(t)
	testutil.UseFakeKeyring(t)
	store, _ := secrets.NewStore(s.SecretsPath)

	reg := evalruntime.DefaultRegistry(store)
	// Mock needs no external deps - its DoctorCheck must report Available
	// so CI boxes with no agent tooling can still run evals end-to-end.
	mock, err := reg.ByName("mock")
	if err != nil {
		t.Fatalf("ByName(mock): %v", err)
	}
	if !mock.DoctorCheck(context.Background()).Available {
		t.Error("mock runner DoctorCheck reported Available=false")
	}
}
