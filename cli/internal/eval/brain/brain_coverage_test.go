package brain_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/brain"
)

func writeBrainTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for p, body := range files {
		full := filepath.Join(root, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSnapshot_NonSmartSkillIsNoOp(t *testing.T) {
	skill := t.TempDir()
	// No references/ directory → not a smart skill.
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("# x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(t.TempDir(), "snap")
	if err := brain.Snapshot(skill, dst); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	if _, err := os.Stat(dst); err == nil {
		t.Error("snapshot should not create dst for non-smart skill")
	}
}

func TestSnapshot_CopiesReferencesTree(t *testing.T) {
	skill := t.TempDir()
	writeBrainTree(t, skill, map[string]string{
		"SKILL.md":                      "# s",
		"references/patterns.md":        "### p1\nbody\n",
		"references/wiki/concept/one.md": "# one",
		"references/raw/note.txt":       "raw",
	})
	dst := filepath.Join(t.TempDir(), "snap")
	if err := brain.Snapshot(skill, dst); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "patterns.md")); err != nil {
		t.Error("patterns.md not copied")
	}
	if _, err := os.Stat(filepath.Join(dst, "wiki", "concept", "one.md")); err != nil {
		t.Error("wiki tree not copied")
	}
}

func TestRestore_ReplacesReferencesDir(t *testing.T) {
	// Seed a snapshot with content.
	snap := t.TempDir()
	if err := os.WriteFile(filepath.Join(snap, "patterns.md"), []byte("### restored"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Skill with OLD content in references/ that should get replaced.
	skill := t.TempDir()
	if err := os.MkdirAll(filepath.Join(skill, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill, "references", "stale.md"), []byte("OLD"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := brain.Restore(snap, skill); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if _, err := os.Stat(filepath.Join(skill, "references", "stale.md")); err == nil {
		t.Error("stale.md should be gone after Restore")
	}
	got, err := os.ReadFile(filepath.Join(skill, "references", "patterns.md"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(got), "restored") {
		t.Errorf("got %q", got)
	}
}

func TestDeriveFlat_ShapesMetaAndStripsWikiRaw(t *testing.T) {
	src := t.TempDir()
	writeBrainTree(t, src, map[string]string{
		"SKILL.md":                       "# foo\n",
		"scripts/lint.sh":                "#!/bin/sh",
		"references/_brain.md":           "brain spec",
		"references/_template.md":        "template",
		"references/patterns.md":         "# patterns\n\n---\n\n### entry one\nbody\n",
		"references/decisions.md":        "# decisions\n\nheader paragraph only\n\n### d1\nbody\n",
		"references/log.md":              "# log\n\n[RUN 2026-01-01] stuff\n",
		"references/_index.md":           "index",
		"references/wiki/concept/one.md": "concept",
		"references/raw/note.txt":        "raw content",
	})

	dst := filepath.Join(t.TempDir(), "flat")
	got, err := brain.DeriveFlat(src, dst)
	if err != nil {
		t.Fatalf("DeriveFlat: %v", err)
	}
	if got != dst {
		t.Errorf("got %q, want %q", got, dst)
	}
	// SKILL.md + scripts + structural brain files retained.
	for _, keep := range []string{
		"SKILL.md", "scripts/lint.sh", "references/_brain.md", "references/_template.md",
	} {
		if _, err := os.Stat(filepath.Join(dst, filepath.FromSlash(keep))); err != nil {
			t.Errorf("missing %s after DeriveFlat", keep)
		}
	}
	// wiki/ and raw/ must be absent.
	if _, err := os.Stat(filepath.Join(dst, "references", "wiki")); err == nil {
		t.Error("wiki/ should be stripped")
	}
	if _, err := os.Stat(filepath.Join(dst, "references", "raw")); err == nil {
		t.Error("raw/ should be stripped")
	}
	// Shaped meta: patterns.md truncated at `---` and has the flat arm note.
	if data, err := os.ReadFile(filepath.Join(dst, "references", "patterns.md")); err != nil {
		t.Fatalf("patterns.md: %v", err)
	} else if !strings.Contains(string(data), "no entries yet") {
		t.Errorf("patterns.md not shaped: %q", data)
	}
}

func TestSHA_DeterministicAndContentSensitive(t *testing.T) {
	a := t.TempDir()
	writeBrainTree(t, a, map[string]string{"a.md": "alpha", "dir/b.md": "beta"})
	shaA, err := brain.SHA(a)
	if err != nil {
		t.Fatalf("SHA: %v", err)
	}

	b := t.TempDir()
	writeBrainTree(t, b, map[string]string{"a.md": "alpha", "dir/b.md": "beta"})
	shaB, err := brain.SHA(b)
	if err != nil {
		t.Fatalf("SHA: %v", err)
	}
	if shaA != shaB {
		t.Errorf("identical trees produced different SHAs: %s vs %s", shaA, shaB)
	}

	c := t.TempDir()
	writeBrainTree(t, c, map[string]string{"a.md": "ALPHA", "dir/b.md": "beta"})
	shaC, _ := brain.SHA(c)
	if shaC == shaA {
		t.Error("different content produced same SHA")
	}
}

func TestComputeGrowth_CountsDeltas(t *testing.T) {
	before := t.TempDir()
	writeBrainTree(t, before, map[string]string{
		"patterns.md":           "### a\n### b\n",
		"wiki/one.md":           "one",
	})
	after := t.TempDir()
	writeBrainTree(t, after, map[string]string{
		"patterns.md":           "### a\n### b\n### c\n",
		"wiki/one.md":           "one",
		"wiki/two.md":           "two",
		"raw/note.md":           "raw",
		"log.md":                "[RUN 2026-01-01] x\n[INGEST 2026-01-02] y\n",
	})
	g, err := brain.ComputeGrowth(before, after)
	if err != nil {
		t.Fatalf("ComputeGrowth: %v", err)
	}
	if g.PatternsEntries.Total != 3 || g.PatternsEntries.Delta != 1 {
		t.Errorf("PatternsEntries = %+v, want {3,1}", g.PatternsEntries)
	}
	if g.WikiConcepts.Total != 2 || g.WikiConcepts.Delta != 1 {
		t.Errorf("WikiConcepts = %+v", g.WikiConcepts)
	}
	if g.RawFiles.Total != 1 || g.RawFiles.Delta != 1 {
		t.Errorf("RawFiles = %+v", g.RawFiles)
	}
	if g.LogEntries.Total != 2 {
		t.Errorf("LogEntries.Total = %d, want 2", g.LogEntries.Total)
	}
}

func TestComputeGrowth_EmptyBeforeIsFreshBrain(t *testing.T) {
	after := t.TempDir()
	writeBrainTree(t, after, map[string]string{
		"wiki/one.md": "one",
	})
	g, err := brain.ComputeGrowth("", after)
	if err != nil {
		t.Fatalf("ComputeGrowth: %v", err)
	}
	if g.WikiConcepts.Total != 1 || g.WikiConcepts.Delta != 1 {
		t.Errorf("WikiConcepts = %+v", g.WikiConcepts)
	}
}

func TestComputeGrowth_NonExistentBeforeDirTreatedAsEmpty(t *testing.T) {
	after := t.TempDir()
	writeBrainTree(t, after, map[string]string{"patterns.md": "### a\n"})
	g, err := brain.ComputeGrowth(filepath.Join(t.TempDir(), "ghost"), after)
	if err != nil {
		t.Fatalf("ComputeGrowth: %v", err)
	}
	if g.PatternsEntries.Total != 1 {
		t.Errorf("got %d", g.PatternsEntries.Total)
	}
}

func TestReadsFromBrain_CountsReferencesHits(t *testing.T) {
	transcript := []byte(`
Read: references/patterns.md
Invoked tool: Read on references/decisions.md
write: references/log.md (non-read, should not count)
Read: somewhere/else (no references, should not count)
read references/_brain.md
`)
	got := brain.ReadsFromBrain(transcript)
	// Three lines contain both "references/" and "Read"/"read".
	if got < 3 {
		t.Errorf("got %d, want >=3", got)
	}
}

func TestReadsFromBrain_EmptyTranscript(t *testing.T) {
	if got := brain.ReadsFromBrain(nil); got != 0 {
		t.Errorf("got %d", got)
	}
	if got := brain.ReadsFromBrain([]byte("")); got != 0 {
		t.Errorf("got %d", got)
	}
}

func TestGrowth_MarshalJSON_RoundTrips(t *testing.T) {
	g := &brain.Growth{
		WikiConcepts: brain.Pair{Total: 5, Delta: 1},
		LogEntries:   brain.Pair{Total: 10, Delta: 2},
	}
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var round brain.Growth
	if err := json.Unmarshal(data, &round); err != nil {
		t.Fatal(err)
	}
	if round.WikiConcepts.Total != 5 || round.LogEntries.Delta != 2 {
		t.Errorf("round-trip lost data: %+v", round)
	}
}

func TestShape_Behavior(t *testing.T) {
	// Direct shape is unexported; exercise it via DeriveFlat on
	// crafted input that hits each of its three branches.
	src := t.TempDir()
	writeBrainTree(t, src, map[string]string{
		"SKILL.md":              "# s\n",
		"references/_brain.md":  "x",
		// Only first paragraph retained (no --- separator).
		"references/patterns.md": "first para\n\nsecond para should vanish",
		// Too short to shape — whole body retained.
		"references/log.md":      "only",
	})
	dst := filepath.Join(t.TempDir(), "flat")
	if _, err := brain.DeriveFlat(src, dst); err != nil {
		t.Fatalf("DeriveFlat: %v", err)
	}
	p, _ := os.ReadFile(filepath.Join(dst, "references", "patterns.md"))
	if strings.Contains(string(p), "second para") {
		t.Errorf("second paragraph should be dropped: %q", p)
	}
}
