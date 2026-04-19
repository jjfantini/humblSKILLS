package brain

import (
	"os"
	"path/filepath"
	"testing"
)

func writeF(t *testing.T, path, body string) {
	t.Helper()
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSnapshotRestoreRoundtrip(t *testing.T) {
	skill := t.TempDir()
	writeF(t, filepath.Join(skill, "SKILL.md"), "name: demo\n")
	writeF(t, filepath.Join(skill, "references", "log.md"), "# log\n---\n[INGEST 2026-04-01] hi\n")
	writeF(t, filepath.Join(skill, "references", "wiki", "ctx", "cat", "c.md"), "x")

	snap := filepath.Join(t.TempDir(), "snap")
	if err := Snapshot(skill, snap); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	// Wipe the live brain and restore.
	if err := os.RemoveAll(filepath.Join(skill, "references")); err != nil {
		t.Fatal(err)
	}
	if err := Restore(snap, skill); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if _, err := os.Stat(filepath.Join(skill, "references", "log.md")); err != nil {
		t.Fatalf("log.md did not restore: %v", err)
	}
	if _, err := os.Stat(filepath.Join(skill, "references", "wiki", "ctx", "cat", "c.md")); err != nil {
		t.Fatalf("wiki did not restore: %v", err)
	}
}

func TestDeriveFlatStripsBrainData(t *testing.T) {
	src := t.TempDir()
	writeF(t, filepath.Join(src, "SKILL.md"), "name: demo\n")
	writeF(t, filepath.Join(src, "scripts", "lint.sh"), "#!/usr/bin/env bash\n")
	writeF(t, filepath.Join(src, "references", "_brain.md"), "spec")
	writeF(t, filepath.Join(src, "references", "log.md"),
		"# Log\nshape docs\n---\n[INGEST 2026-04-01] entry one\n[INGEST 2026-04-02] entry two\n")
	writeF(t, filepath.Join(src, "references", "wiki", "a", "b", "c.md"), "x")
	writeF(t, filepath.Join(src, "references", "raw", "note.md"), "x")

	dst := filepath.Join(t.TempDir(), "flat")
	if _, err := DeriveFlat(src, dst); err != nil {
		t.Fatalf("DeriveFlat: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "SKILL.md")); err != nil {
		t.Fatalf("SKILL.md missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "scripts", "lint.sh")); err != nil {
		t.Fatalf("scripts/lint.sh missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "references", "wiki")); err == nil {
		t.Fatalf("wiki/ should be removed in flat skill")
	}
	if _, err := os.Stat(filepath.Join(dst, "references", "raw")); err == nil {
		t.Fatalf("raw/ should be removed in flat skill")
	}
	body, err := os.ReadFile(filepath.Join(dst, "references", "log.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) == "" {
		t.Fatalf("log.md should retain its shape preamble")
	}
	if countLogEntries(filepath.Join(dst, "references", "log.md")) != 0 {
		t.Fatalf("log entries should be stripped")
	}
}

func TestComputeGrowth(t *testing.T) {
	before := t.TempDir()
	after := t.TempDir()
	writeF(t, filepath.Join(before, "wiki", "a", "b", "one.md"), "x")
	writeF(t, filepath.Join(after, "wiki", "a", "b", "one.md"), "x")
	writeF(t, filepath.Join(after, "wiki", "a", "b", "two.md"), "xx")
	writeF(t, filepath.Join(after, "raw", "note.md"), "x")
	writeF(t, filepath.Join(after, "log.md"), "# Log\n---\n[INGEST 2026-04-01] a\n[INGEST 2026-04-02] b\n")

	g, err := ComputeGrowth(before, after)
	if err != nil {
		t.Fatalf("ComputeGrowth: %v", err)
	}
	if g.WikiConcepts.Total != 2 || g.WikiConcepts.Delta != 1 {
		t.Fatalf("wiki concepts: %+v", g.WikiConcepts)
	}
	if g.RawFiles.Delta != 1 {
		t.Fatalf("raw files delta: %d", g.RawFiles.Delta)
	}
	if g.LogEntries.Total != 2 {
		t.Fatalf("log entries total: %d", g.LogEntries.Total)
	}
}

func TestReadsFromBrain(t *testing.T) {
	trans := []byte(
		"[mock] Read: references/patterns.md\n" +
			"[mock] Write: out.md\n" +
			"Read: references/decisions.md\n" +
			"Read: some-other-file.md\n")
	if got := ReadsFromBrain(trans); got != 2 {
		t.Fatalf("expected 2 brain reads, got %d", got)
	}
}
