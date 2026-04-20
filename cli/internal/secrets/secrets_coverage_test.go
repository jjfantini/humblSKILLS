package secrets_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/secrets"
	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

// Tests in this file use the testutil sandbox so XDG resolution hits
// sandboxed paths, letting us exercise NewStore's default-path branch
// and DefaultFilePath without risk of polluting the developer's home.

func TestNewStore_DefaultsToXDGConfig(t *testing.T) {
	s := testutil.NewSandbox(t)
	testutil.UseFakeKeyring(t)

	store, err := secrets.NewStore("")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	// Driving Set + Get confirms the default path is usable.
	if _, err := store.Set("anthropic", "k"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	// File may or may not exist depending on whether the mock keyring
	// accepted the value, but DefaultFilePath must point inside the
	// sandbox XDG config tree.
	got, err := secrets.DefaultFilePath()
	if err != nil {
		t.Fatalf("DefaultFilePath: %v", err)
	}
	if !strings.HasPrefix(got, s.XDGConfigHome) {
		t.Errorf("DefaultFilePath = %q, want prefix %q", got, s.XDGConfigHome)
	}
	if filepath.Base(got) != "secrets.json" {
		t.Errorf("DefaultFilePath basename = %q", filepath.Base(got))
	}
}

func TestSet_UnknownProvider(t *testing.T) {
	s := testutil.NewSandbox(t)
	testutil.UseFakeKeyring(t)
	store, _ := secrets.NewStore(s.SecretsPath)
	if _, err := store.Set("nope", "value"); err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestSet_EmptyValueRefused(t *testing.T) {
	s := testutil.NewSandbox(t)
	testutil.UseFakeKeyring(t)
	store, _ := secrets.NewStore(s.SecretsPath)
	if _, err := store.Set("anthropic", "   "); err == nil {
		t.Fatal("expected error when storing empty/whitespace secret")
	}
}

func TestSet_KeyringUnavailable_WithFileFallbackFailure(t *testing.T) {
	s := testutil.NewSandbox(t)
	testutil.UseUnavailableKeyring(t, errors.New("dbus down"))

	// Route the file path through a location that can't be created
	// (an existing regular file at what should be the secrets dir).
	blockPath := filepath.Join(s.Root, "blocker")
	if err := os.WriteFile(blockPath, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	badSecretsPath := filepath.Join(blockPath, "secrets.json") // parent is a file!

	store, err := secrets.NewStore(badSecretsPath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	_, err = store.Set("anthropic", "should-fail")
	if err == nil {
		t.Fatal("expected error when both keyring and file fallback fail")
	}
	if !strings.Contains(err.Error(), "keyring unavailable") {
		t.Errorf("err message should mention keyring unavailable: %v", err)
	}
}

func TestGet_UnknownProviderErrors(t *testing.T) {
	s := testutil.NewSandbox(t)
	testutil.UseFakeKeyring(t)
	store, _ := secrets.NewStore(s.SecretsPath)
	_, _, err := store.Get("ghost")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestDelete_UnknownProviderErrors(t *testing.T) {
	s := testutil.NewSandbox(t)
	testutil.UseFakeKeyring(t)
	store, _ := secrets.NewStore(s.SecretsPath)
	if err := store.Delete("ghost"); err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestSet_RemovesStaleFileCopyOnKeyringSuccess(t *testing.T) {
	s := testutil.NewSandbox(t)
	testutil.UseFakeKeyring(t)
	store, _ := secrets.NewStore(s.SecretsPath)

	// Unavailable keyring puts the secret in the file...
	testutil.UseUnavailableKeyring(t, errors.New("gone"))
	if _, err := store.Set("anthropic", "from-file"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if _, err := os.Stat(s.SecretsPath); err != nil {
		t.Fatalf("file should exist after file-fallback Set: %v", err)
	}

	// ...then keyring comes back online; a new Set should land in the
	// keyring AND remove the stale file entry. writeFileSecret("") on
	// the last key removes the whole file.
	testutil.UseFakeKeyring(t)
	if _, err := store.Set("anthropic", "from-keyring"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if _, err := os.Stat(s.SecretsPath); !os.IsNotExist(err) {
		t.Errorf("stale file should be removed once keyring succeeds; stat err=%v", err)
	}
}

func TestKeyringAvailable_WithMockInit(t *testing.T) {
	testutil.UseFakeKeyring(t)
	if !secrets.KeyringAvailable() {
		t.Error("mock keyring should report Available=true")
	}
}

func TestKeyringAvailable_WithBrokenKeyring(t *testing.T) {
	testutil.UseUnavailableKeyring(t, errors.New("service missing"))
	if secrets.KeyringAvailable() {
		t.Error("broken keyring should report Available=false")
	}
}

func TestReadFile_IgnoresCorruptJSON(t *testing.T) {
	s := testutil.NewSandbox(t)
	// Keyring empty so Get must fall through to the file layer.
	testutil.UseFakeKeyring(t)
	if err := os.MkdirAll(filepath.Dir(s.SecretsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.SecretsPath, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	store, _ := secrets.NewStore(s.SecretsPath)
	_, src, err := store.Get("anthropic")
	if err != nil {
		t.Fatalf("Get on corrupt file returned err: %v", err)
	}
	if src != secrets.SourceAbsent {
		t.Errorf("source = %q, want absent (corrupt file treated as missing)", src)
	}
}

func TestProviders_IsStableSorted(t *testing.T) {
	ps := secrets.Providers()
	if len(ps) == 0 {
		t.Fatal("Providers returned empty list")
	}
	for i := 1; i < len(ps); i++ {
		if ps[i].Name < ps[i-1].Name {
			t.Errorf("Providers not sorted: %v", ps)
			break
		}
	}
}

func TestFilePerm_IsHonoredOnPOSIX(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits not honored on windows")
	}
	s := testutil.NewSandbox(t)
	testutil.UseUnavailableKeyring(t, errors.New("force file path"))
	store, _ := secrets.NewStore(s.SecretsPath)
	if _, err := store.Set("openai", "sk-file"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	info, err := os.Stat(s.SecretsPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("perm = %o, want 0600", info.Mode().Perm())
	}
}
