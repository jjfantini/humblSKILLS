package secrets

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zalando/go-keyring"
)

// Use the mock keyring so tests are hermetic - never touches the real OS
// keychain during `go test`.
func init() { keyring.MockInit() }

func TestEnvBeatsStoredValues(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")
	dir := t.TempDir()
	store, err := NewStore(filepath.Join(dir, "secrets.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	// Set via the store (keyring in mock mode).
	if _, err := store.Set("anthropic", "stored-key"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	t.Setenv("ANTHROPIC_API_KEY", "env-key")
	v, src, err := store.Get("anthropic")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if v != "env-key" || src != SourceEnv {
		t.Fatalf("got %q/%s, want env-key/env", v, src)
	}
}

func TestFileFallbackWhenKeyringMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.json")
	if err := writeFileSecret(path, "anthropic", "file-key"); err != nil {
		t.Fatalf("writeFileSecret: %v", err)
	}
	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	// No env, no keyring value for this provider. But the mock keyring
	// holds whatever we set in the previous test; reset by using a fresh
	// provider that hasn't been touched.
	os.Unsetenv("ANTHROPIC_API_KEY")
	// Ensure keyring mock has nothing for this provider.
	_ = store.Delete("anthropic")
	// Restore file copy (Delete wipes it).
	if err := writeFileSecret(path, "anthropic", "file-key"); err != nil {
		t.Fatalf("writeFileSecret: %v", err)
	}
	v, src, err := store.Get("anthropic")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if v != "file-key" || src != SourceFile {
		t.Fatalf("got %q/%s, want file-key/file", v, src)
	}
	// Permissions on the file should be 0600.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("got perm %v, want 0600", info.Mode().Perm())
	}
}

func TestDeleteWipesEverywhere(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")
	dir := t.TempDir()
	store, err := NewStore(filepath.Join(dir, "secrets.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if _, err := store.Set("openai", "value"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := store.Delete("openai"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, src, err := store.Get("openai")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if src != SourceAbsent {
		t.Fatalf("expected absent, got %s", src)
	}
}

func TestProviderByName(t *testing.T) {
	if _, ok := ProviderByName("anthropic"); !ok {
		t.Fatalf("anthropic should be known")
	}
	if _, ok := ProviderByName("nope"); ok {
		t.Fatalf("nope should be unknown")
	}
}
