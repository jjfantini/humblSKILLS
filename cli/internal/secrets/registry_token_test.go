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

func TestRegistryToken_NamedAccounts(t *testing.T) {
	t.Setenv(RegistryTokenEnvVar, "")
	path := filepath.Join(t.TempDir(), "secrets.json")

	// Store a token for the "work" registry.
	if _, err := setKeyedToken(path, registryAccount("work"), "work-tok"); err != nil {
		t.Fatal(err)
	}
	if v, src := getKeyedToken(path, registryAccount("work")); v != "work-tok" || src == SourceAbsent {
		t.Fatalf("work token = (%q,%s), want (work-tok, present)", v, src)
	}
	// A different name's account is isolated (absent).
	if v, src := getKeyedToken(path, registryAccount("public")); v != "" || src != SourceAbsent {
		t.Fatalf("public token = (%q,%s), want (\"\", absent)", v, src)
	}
	// The default account is also isolated from named ones.
	if v, src := getKeyedToken(path, registryAccount("")); v != "" || src != SourceAbsent {
		t.Fatalf("default token = (%q,%s), want (\"\", absent)", v, src)
	}

	if err := deleteKeyedToken(path, registryAccount("work")); err != nil {
		t.Fatal(err)
	}
	if _, src := getKeyedToken(path, registryAccount("work")); src != SourceAbsent {
		t.Fatal("work token should be absent after delete")
	}
}

func TestRegistryToken_AccountNaming(t *testing.T) {
	if got := registryAccount(""); got != "registry-token" {
		t.Errorf("default account = %q, want registry-token", got)
	}
	if got := registryAccount("work"); got != "registry-token-work" {
		t.Errorf("named account = %q, want registry-token-work", got)
	}
}

func TestRegistryToken_KeyringAccountIsolated(t *testing.T) {
	// The registry token must not collide with an LLM provider's "-api-key"
	// keyring account.
	if got := registryTokenKey; got == keyringAccount("anthropic") || got == keyringAccount("openai") {
		t.Fatalf("registry token key %q collides with a provider api-key account", got)
	}
}
