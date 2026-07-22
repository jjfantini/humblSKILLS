package profile

import (
	"path/filepath"
	"testing"
)

func TestNamedRegistries_SetFindRemove(t *testing.T) {
	p := &Profile{}
	p.SetRegistry("public", "https://a")
	p.SetRegistry("work", "https://b")
	if len(p.Registries) != 2 {
		t.Fatalf("want 2 registries, got %d", len(p.Registries))
	}

	// SetRegistry with an existing name updates in place (no duplicate).
	p.SetRegistry("public", "https://a2")
	if r, ok := p.FindRegistry("public"); !ok || r.URL != "https://a2" {
		t.Fatalf("update failed: %+v (ok=%v)", r, ok)
	}
	if len(p.Registries) != 2 {
		t.Fatalf("update must not add; got %d", len(p.Registries))
	}

	if _, ok := p.FindRegistry("missing"); ok {
		t.Fatal("FindRegistry found a nonexistent name")
	}

	if !p.RemoveRegistry("public") {
		t.Fatal("RemoveRegistry should report true for an existing name")
	}
	if _, ok := p.FindRegistry("public"); ok {
		t.Fatal("registry still present after remove")
	}
	if p.RemoveRegistry("public") {
		t.Fatal("second remove should report false")
	}
	if len(p.Registries) != 1 {
		t.Fatalf("want 1 registry after remove, got %d", len(p.Registries))
	}
}

func TestNamedRegistries_Rename(t *testing.T) {
	p := &Profile{}
	p.SetRegistry("old", "https://a")
	p.SetRegistry("keep", "https://b")

	if err := p.RenameRegistry("old", "new"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if _, ok := p.FindRegistry("new"); !ok {
		t.Fatal("renamed registry missing under new name")
	}
	if _, ok := p.FindRegistry("old"); ok {
		t.Fatal("old name still present after rename")
	}
	// Collision with an existing name is rejected.
	if err := p.RenameRegistry("new", "keep"); err == nil {
		t.Fatal("expected error renaming onto an existing name")
	}
	// Renaming a missing registry is rejected.
	if err := p.RenameRegistry("missing", "x"); err == nil {
		t.Fatal("expected error renaming a nonexistent registry")
	}
}

func TestNamedRegistries_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.json")
	p := &Profile{SchemaVersion: SchemaVersion}
	p.SetRegistry("work", "https://x")
	if err := Save(path, p); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if r, ok := got.FindRegistry("work"); !ok || r.URL != "https://x" {
		t.Fatalf("round-trip failed: %+v", got.Registries)
	}
}
