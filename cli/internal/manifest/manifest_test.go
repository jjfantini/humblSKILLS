package manifest

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_MissingReturnsEmpty(t *testing.T) {
	p := filepath.Join(t.TempDir(), "manifest.json")
	m, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if m.SchemaVersion != SchemaVersion {
		t.Errorf("got schema %d", m.SchemaVersion)
	}
	if len(m.Installations) != 0 {
		t.Errorf("got %d installations", len(m.Installations))
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "nested", "manifest.json")
	want := &Manifest{
		SchemaVersion: SchemaVersion,
		Installations: []Installation{{
			Skill:       "foo",
			Version:     "0.1.0",
			Platform:    "claude-code",
			Scope:       "user",
			Path:        "/tmp/foo",
			InstalledAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			SourceSHA:   "abc",
			RegistryRef: "def",
		}},
	}
	if err := Save(p, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Installations) != 1 || got.Installations[0].Skill != "foo" {
		t.Errorf("round-trip mismatch: %+v", got)
	}
	if !got.Installations[0].InstalledAt.Equal(want.Installations[0].InstalledAt) {
		t.Errorf("time mismatch: %v vs %v", got.Installations[0].InstalledAt, want.Installations[0].InstalledAt)
	}
}

func TestSave_Atomic(t *testing.T) {
	p := filepath.Join(t.TempDir(), "manifest.json")
	if err := Save(p, &Manifest{}); err != nil {
		t.Fatal(err)
	}
	// Tmp file should be gone.
	if _, err := os.Stat(p + ".tmp"); err == nil {
		t.Error("expected tmp file to be cleaned up")
	}
}

func TestLoad_RejectsFutureSchema(t *testing.T) {
	p := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(p, []byte(`{"schema_version":999}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p); err == nil {
		t.Fatal("expected schema mismatch error")
	}
}

func TestFind(t *testing.T) {
	m := &Manifest{Installations: []Installation{{Skill: "a"}, {Skill: "b"}}}
	if got := m.Find("a"); got == nil || got.Skill != "a" {
		t.Errorf("got %+v", got)
	}
	if got := m.Find("ghost"); got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestFindAllFindOneUpsertRemove(t *testing.T) {
	m := &Manifest{}
	m.Upsert(Installation{Skill: "a", Platform: "claude-code", Scope: "user", Version: "0.1.0"})
	m.Upsert(Installation{Skill: "a", Platform: "cursor", Scope: "user", Version: "0.1.0"})
	m.Upsert(Installation{Skill: "b", Platform: "claude-code", Scope: "project"})

	if all := m.FindAll("a"); len(all) != 2 {
		t.Fatalf("FindAll(a) = %d", len(all))
	}
	if e := m.FindOne("a", "claude-code", "user"); e == nil || e.Version != "0.1.0" {
		t.Fatalf("FindOne missed: %+v", e)
	}
	// Upsert replaces instead of appending.
	m.Upsert(Installation{Skill: "a", Platform: "claude-code", Scope: "user", Version: "0.2.0"})
	if len(m.Installations) != 3 {
		t.Errorf("upsert duplicated: %d", len(m.Installations))
	}
	if e := m.FindOne("a", "claude-code", "user"); e == nil || e.Version != "0.2.0" {
		t.Errorf("upsert didn't replace: %+v", e)
	}

	if !m.RemoveOne("a", "cursor", "user") {
		t.Error("RemoveOne returned false")
	}
	if got := m.Remove("a"); got != 1 {
		t.Errorf("Remove(a) = %d", got)
	}
	if len(m.Installations) != 1 || m.Installations[0].Skill != "b" {
		t.Errorf("unexpected state: %+v", m.Installations)
	}
}

func TestDefaultPath_NonEmpty(t *testing.T) {
	p, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if p == "" {
		t.Error("empty default path")
	}
}
