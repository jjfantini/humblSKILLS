package secrets

import (
	"path/filepath"
	"testing"
)

// keyring is mocked in-process by init() in secrets_test.go, so Set/Get/Delete
// operate on an in-memory store and never touch the real OS keychain.

func TestRegistryToken_EnvWins(t *testing.T) {
	t.Setenv(RegistryTokenEnvVar, "env-token")
	// Even with a keyring value present, env takes precedence.
	if _, err := setRegistryToken(filepath.Join(t.TempDir(), "secrets.json"), "keyring-token"); err != nil {
		t.Fatal(err)
	}
	v, src := getRegistryToken("")
	if v != "env-token" || src != SourceEnv {
		t.Fatalf("got (%q,%s), want (env-token,env)", v, src)
	}
	_ = deleteRegistryToken(filepath.Join(t.TempDir(), "secrets.json"))
}

func TestRegistryToken_KeyringRoundTrip(t *testing.T) {
	t.Setenv(RegistryTokenEnvVar, "") // ensure env doesn't shadow
	path := filepath.Join(t.TempDir(), "secrets.json")

	src, err := setRegistryToken(path, "  ghp_secret  ") // also checks trimming
	if err != nil {
		t.Fatal(err)
	}
	if src != SourceKeyring {
		t.Fatalf("Set source = %s, want keyring (mock)", src)
	}

	v, gsrc := getRegistryToken(path)
	if v != "ghp_secret" || gsrc != SourceKeyring {
		t.Fatalf("Get = (%q,%s), want (ghp_secret,keyring)", v, gsrc)
	}

	if err := deleteRegistryToken(path); err != nil {
		t.Fatal(err)
	}
	if v, src := getRegistryToken(path); v != "" || src != SourceAbsent {
		t.Fatalf("after delete Get = (%q,%s), want (\"\",absent)", v, src)
	}
}

func TestRegistryToken_RejectsEmpty(t *testing.T) {
	if _, err := setRegistryToken(filepath.Join(t.TempDir(), "secrets.json"), "   "); err == nil {
		t.Fatal("expected error storing empty token")
	}
}

func TestRegistryToken_KeyringAccountIsolated(t *testing.T) {
	// The registry token must not collide with an LLM provider's "-api-key"
	// keyring account.
	if got := registryTokenKey; got == keyringAccount("anthropic") || got == keyringAccount("openai") {
		t.Fatalf("registry token key %q collides with a provider api-key account", got)
	}
}
