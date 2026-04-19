package adapters

import (
	"path/filepath"
	"testing"
)

func TestIsWritable_Yes(t *testing.T) {
	if !isWritable(t.TempDir()) {
		t.Error("tempdir should be writable")
	}
}

func TestIsWritable_NonExistentParent(t *testing.T) {
	dir := t.TempDir()
	// Nested missing path - isWritable should walk up and probe the real parent.
	nested := filepath.Join(dir, "a", "b", "c")
	if !isWritable(nested) {
		t.Error("nested missing path under writable parent should report writable")
	}
}

func TestTargets_Sorted(t *testing.T) {
	a := Adapter{
		Name: "fake",
		InstallTargets: map[string]string{
			"user":    "/tmp/x/user",
			"project": "/tmp/x/project",
		},
	}
	got := a.Targets()
	if len(got) != 2 || got[0].Scope != "project" || got[1].Scope != "user" {
		t.Errorf("expected project, user order, got %+v", got)
	}
}

func TestTarget_DefaultScope(t *testing.T) {
	a := Adapter{
		Name:           "fake",
		DefaultScope:   "user",
		InstallTargets: map[string]string{"user": "/tmp/ok"},
	}
	tgt, err := a.Target("")
	if err != nil {
		t.Fatal(err)
	}
	if tgt.Scope != "user" {
		t.Errorf("got %q", tgt.Scope)
	}
}

func TestTarget_UnknownScope(t *testing.T) {
	a := Adapter{Name: "fake", InstallTargets: map[string]string{"user": "/tmp"}}
	if _, err := a.Target("project"); err == nil {
		t.Error("expected error for unknown scope")
	}
}
