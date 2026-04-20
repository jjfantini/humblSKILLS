package workspace

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIterationLifecycle(t *testing.T) {
	root := t.TempDir()
	skill := "demo"

	n1, dir1, err := BeginIteration(root, skill, "mock", []string{"smart_skill"}, []string{"s1"})
	if err != nil {
		t.Fatalf("BeginIteration: %v", err)
	}
	if n1 != 1 {
		t.Fatalf("first iteration should be 1, got %d", n1)
	}
	if _, err := os.Stat(dir1); err != nil {
		t.Fatalf("dir not created: %v", err)
	}
	if err := CompleteIteration(root, skill, n1,
		map[string]float64{"smart_skill": 0.9}, map[string]int{"smart_skill": 1200}); err != nil {
		t.Fatalf("CompleteIteration: %v", err)
	}

	n2, _, err := BeginIteration(root, skill, "mock", []string{"smart_skill"}, []string{"s1"})
	if err != nil {
		t.Fatalf("BeginIteration 2: %v", err)
	}
	if n2 != 2 {
		t.Fatalf("second iteration should be 2, got %d", n2)
	}

	reg, err := LoadRegistry(root, skill)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	if len(reg.Iterations) != 2 {
		t.Fatalf("expected 2 iterations, got %d", len(reg.Iterations))
	}
	if reg.Iterations[0].Status != StatusComplete {
		t.Fatalf("expected first iteration complete, got %s", reg.Iterations[0].Status)
	}
}

func TestPruneKeepLast(t *testing.T) {
	root := t.TempDir()
	skill := "demo"
	// Seed 5 iterations all marked complete.
	for i := 0; i < 5; i++ {
		n, _, err := BeginIteration(root, skill, "mock", nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := CompleteIteration(root, skill, n, nil, nil); err != nil {
			t.Fatal(err)
		}
	}
	res, err := Prune(root, skill, PruneOpts{KeepLast: 2})
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if len(res.Removed) != 3 {
		t.Fatalf("expected 3 removed, got %v", res.Removed)
	}
	reg, _ := LoadRegistry(root, skill)
	if len(reg.Iterations) != 2 {
		t.Fatalf("expected 2 remaining, got %d", len(reg.Iterations))
	}
	// Check that iteration directories are gone.
	for _, n := range res.Removed {
		if _, err := os.Stat(IterationDir(root, skill, n)); !os.IsNotExist(err) {
			t.Fatalf("iteration %d still on disk: %v", n, err)
		}
	}
}

func TestPruneOlderThan(t *testing.T) {
	root := t.TempDir()
	skill := "demo"
	n1, _, _ := BeginIteration(root, skill, "mock", nil, nil)
	_ = CompleteIteration(root, skill, n1, nil, nil)

	// Backdate iteration 1 by 40 days.
	reg, _ := LoadRegistry(root, skill)
	reg.Iterations[0].StartedAt = time.Now().Add(-40 * 24 * time.Hour)
	_ = SaveRegistry(root, skill, reg)

	// Fresh iteration.
	n2, _, _ := BeginIteration(root, skill, "mock", nil, nil)
	_ = CompleteIteration(root, skill, n2, nil, nil)

	res, err := Prune(root, skill, PruneOpts{OlderThan: 30 * 24 * time.Hour})
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if len(res.Removed) != 1 || res.Removed[0] != 1 {
		t.Fatalf("expected only iteration 1 removed, got %v", res.Removed)
	}
}

func TestSizeBytes(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a"), []byte("hello"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "b"), []byte("world!"), 0o644)
	n, err := SizeBytes(dir)
	if err != nil {
		t.Fatalf("SizeBytes: %v", err)
	}
	if n != 11 {
		t.Fatalf("expected 11 bytes, got %d", n)
	}
}

func TestHumanSize(t *testing.T) {
	tests := map[int64]string{0: "0 B", 1023: "1023 B", 1024: "1.0 KiB", 1048576: "1.0 MiB"}
	for in, want := range tests {
		if got := HumanSize(in); got != want {
			t.Errorf("HumanSize(%d) = %s, want %s", in, got, want)
		}
	}
}
