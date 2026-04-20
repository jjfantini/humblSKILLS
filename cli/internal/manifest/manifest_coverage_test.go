package manifest_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

func sampleInstall(skill, platform, scope string) manifest.Installation {
	return manifest.Installation{
		Skill:       skill,
		Version:     "1.0.0",
		Platform:    platform,
		Scope:       scope,
		Path:        "/tmp/" + skill,
		InstalledAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		SourceSHA:   "abc",
		RegistryRef: "sha-" + skill,
	}
}

func TestDefaultPath_ResolvesUnderSandboxedXDG(t *testing.T) {
	s := testutil.NewSandbox(t)
	p, err := manifest.DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if !strings.HasPrefix(p, s.XDGStateHome) {
		t.Errorf("DefaultPath %q not under XDG_STATE_HOME %q", p, s.XDGStateHome)
	}
	if filepath.Base(p) != "manifest.json" {
		t.Errorf("basename = %q", filepath.Base(p))
	}
}

func TestLoad_RejectsUnsupportedSchemaVersion(t *testing.T) {
	s := testutil.NewSandbox(t)
	body := `{"schema_version": 42, "installations": []}` + "\n"
	if err := os.MkdirAll(filepath.Dir(s.ManifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.ManifestPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := manifest.Load(s.ManifestPath); err == nil {
		t.Fatal("expected error for unsupported schema_version")
	}
}

func TestLoad_RejectsCorruptJSON(t *testing.T) {
	s := testutil.NewSandbox(t)
	if err := os.MkdirAll(filepath.Dir(s.ManifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.ManifestPath, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := manifest.Load(s.ManifestPath); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoad_UpgradesLegacyZeroSchemaVersion(t *testing.T) {
	s := testutil.NewSandbox(t)
	body := `{"installations":[]}` + "\n"
	if err := os.MkdirAll(filepath.Dir(s.ManifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.ManifestPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := manifest.Load(s.ManifestPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if m.SchemaVersion != manifest.SchemaVersion {
		t.Errorf("SchemaVersion not upgraded: %d", m.SchemaVersion)
	}
}

func TestSave_NilManifestRejected(t *testing.T) {
	s := testutil.NewSandbox(t)
	if err := manifest.Save(s.ManifestPath, nil); err == nil {
		t.Fatal("expected error for nil manifest")
	}
}

func TestSave_FailsWhenParentUnwritable(t *testing.T) {
	s := testutil.NewSandbox(t)
	parent := filepath.Join(s.Root, "ro")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(parent, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(parent, 0o755) })

	target := filepath.Join(parent, "nested", "manifest.json")
	if err := manifest.Save(target, &manifest.Manifest{}); err == nil {
		t.Fatal("expected Save to fail with unwritable parent")
	}
}

func TestSave_StampsSchemaVersion(t *testing.T) {
	s := testutil.NewSandbox(t)
	m := &manifest.Manifest{} // zero schema_version
	if err := manifest.Save(s.ManifestPath, m); err != nil {
		t.Fatalf("Save: %v", err)
	}
	data, err := os.ReadFile(s.ManifestPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var raw struct {
		SchemaVersion int `json:"schema_version"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if raw.SchemaVersion != manifest.SchemaVersion {
		t.Errorf("schema_version = %d, want %d", raw.SchemaVersion, manifest.SchemaVersion)
	}
}

func TestFind_ReturnsFirstMatch(t *testing.T) {
	m := &manifest.Manifest{SchemaVersion: manifest.SchemaVersion}
	m.Upsert(sampleInstall("foo", "claude-code", "user"))
	m.Upsert(sampleInstall("foo", "cursor-agent", "project"))

	got := m.Find("foo")
	if got == nil {
		t.Fatal("Find returned nil")
	}
	// Order is insertion order — first install wins.
	if got.Platform != "claude-code" {
		t.Errorf("platform = %q", got.Platform)
	}
	if m.Find("absent") != nil {
		t.Error("Find returned non-nil for absent skill")
	}
}

func TestFindAll_ReturnsEveryMatch(t *testing.T) {
	m := &manifest.Manifest{SchemaVersion: manifest.SchemaVersion}
	m.Upsert(sampleInstall("foo", "claude-code", "user"))
	m.Upsert(sampleInstall("foo", "cursor-agent", "project"))
	m.Upsert(sampleInstall("bar", "claude-code", "user"))

	got := m.FindAll("foo")
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

func TestFindOne_PrecisionMatch(t *testing.T) {
	m := &manifest.Manifest{SchemaVersion: manifest.SchemaVersion}
	m.Upsert(sampleInstall("foo", "claude-code", "user"))
	if got := m.FindOne("foo", "claude-code", "user"); got == nil {
		t.Error("FindOne should hit")
	}
	if got := m.FindOne("foo", "claude-code", "project"); got != nil {
		t.Error("FindOne should miss on wrong scope")
	}
	if got := m.FindOne("foo", "cursor-agent", "user"); got != nil {
		t.Error("FindOne should miss on wrong platform")
	}
	if got := m.FindOne("bar", "claude-code", "user"); got != nil {
		t.Error("FindOne should miss on wrong skill")
	}
}

func TestUpsert_ReplacesOnSameKey(t *testing.T) {
	m := &manifest.Manifest{SchemaVersion: manifest.SchemaVersion}
	m.Upsert(sampleInstall("foo", "claude-code", "user"))
	upgraded := sampleInstall("foo", "claude-code", "user")
	upgraded.Version = "2.0.0"
	m.Upsert(upgraded)
	if len(m.Installations) != 1 {
		t.Fatalf("len = %d, want 1", len(m.Installations))
	}
	if m.Installations[0].Version != "2.0.0" {
		t.Errorf("version = %q", m.Installations[0].Version)
	}
}

func TestRemove_DropsEveryMatchAndReturnsCount(t *testing.T) {
	m := &manifest.Manifest{SchemaVersion: manifest.SchemaVersion}
	m.Upsert(sampleInstall("foo", "claude-code", "user"))
	m.Upsert(sampleInstall("foo", "cursor-agent", "project"))
	m.Upsert(sampleInstall("bar", "claude-code", "user"))

	n := m.Remove("foo")
	if n != 2 {
		t.Errorf("Remove returned %d, want 2", n)
	}
	if len(m.Installations) != 1 || m.Installations[0].Skill != "bar" {
		t.Errorf("unexpected remaining: %+v", m.Installations)
	}
	if n := m.Remove("ghost"); n != 0 {
		t.Errorf("Remove ghost returned %d", n)
	}
}

func TestRemoveOne_ExactMatchOnly(t *testing.T) {
	m := &manifest.Manifest{SchemaVersion: manifest.SchemaVersion}
	m.Upsert(sampleInstall("foo", "claude-code", "user"))
	m.Upsert(sampleInstall("foo", "cursor-agent", "project"))

	if !m.RemoveOne("foo", "claude-code", "user") {
		t.Error("RemoveOne should return true on hit")
	}
	if len(m.Installations) != 1 {
		t.Errorf("len = %d, want 1", len(m.Installations))
	}
	if m.RemoveOne("foo", "claude-code", "user") {
		t.Error("second RemoveOne on same key should return false")
	}
}

func TestSaveLoad_RoundTripPreservesEveryField(t *testing.T) {
	s := testutil.NewSandbox(t)
	m := &manifest.Manifest{SchemaVersion: manifest.SchemaVersion}
	m.Upsert(sampleInstall("foo", "claude-code", "user"))
	m.Upsert(sampleInstall("bar", "cursor-agent", "project"))

	if err := manifest.Save(s.ManifestPath, m); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := manifest.Load(s.ManifestPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got.Installations) != 2 {
		t.Fatalf("installs = %d", len(got.Installations))
	}
	for i, want := range m.Installations {
		if got.Installations[i] != want {
			t.Errorf("install[%d] mismatch:\n got=%+v\nwant=%+v", i, got.Installations[i], want)
		}
	}
}
