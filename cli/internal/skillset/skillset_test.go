package skillset

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "humblskills.json")

	s := New()
	s.Add("beta", "2.0.0")
	s.Add("alpha", "1.0.0")
	if err := Save(path, s); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != SchemaVersion {
		t.Errorf("schema = %d", got.SchemaVersion)
	}
	// Save sorts, so alpha comes first.
	if len(got.Skills) != 2 || got.Skills[0].Name != "alpha" || got.Skills[1].Name != "beta" {
		t.Fatalf("unexpected skills: %+v", got.Skills)
	}
	if got.Skills[0].Version != "1.0.0" {
		t.Errorf("alpha version = %q", got.Skills[0].Version)
	}
}

func TestAdd_DedupesLastWins(t *testing.T) {
	s := New()
	s.Add("foo", "1.0.0")
	s.Add("foo", "1.0.1")
	s.Add("  ", "x") // blank name ignored
	if len(s.Skills) != 1 {
		t.Fatalf("skills = %d, want 1", len(s.Skills))
	}
	if s.Skills[0].Version != "1.0.1" {
		t.Errorf("version = %q, want 1.0.1", s.Skills[0].Version)
	}
}

func TestLoad_DefaultsSchemaZero(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s.json")
	if err := os.WriteFile(path, []byte(`{"skills":[{"name":"foo"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("minimal file should load: %v", err)
	}
	if got.SchemaVersion != SchemaVersion || len(got.Skills) != 1 {
		t.Errorf("unexpected: %+v", got)
	}
}

func TestValidate_Errors(t *testing.T) {
	cases := map[string]*Set{
		"bad schema": {SchemaVersion: 999, Skills: []Skill{{Name: "a"}}},
		"empty name": {SchemaVersion: SchemaVersion, Skills: []Skill{{Name: "  "}}},
		"duplicate":  {SchemaVersion: SchemaVersion, Skills: []Skill{{Name: "a"}, {Name: "a"}}},
	}
	for name, s := range cases {
		if err := s.Validate(); err == nil {
			t.Errorf("%s: expected validation error", name)
		}
	}
}

func TestNames(t *testing.T) {
	s := New()
	s.Add("a", "")
	s.Add("b", "")
	names := s.Names()
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Errorf("names = %v", names)
	}
}
