// Package secrets stores API keys for eval runners (Anthropic, OpenAI, ...).
//
// Keys resolve in this order on Get:
//
//  1. provider-standard env var (ANTHROPIC_API_KEY, OPENAI_API_KEY, ...).
//     Env is never persisted.
//  2. OS keyring via zalando/go-keyring (macOS Keychain, Linux Secret Service,
//     Windows Credential Manager). Preferred persistent store.
//  3. File fallback at $XDG_CONFIG_HOME/humblskills/secrets.json (perm 0600)
//     for platforms where the keyring is unavailable.
//
// Set prefers keyring; falls back to file with a loud warning.
package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrg/xdg"
	"github.com/zalando/go-keyring"
)

// Service is the keyring service name. Accounts are "<provider>-api-key".
const Service = "humblskills"

// Source identifies where a key came from.
type Source string

const (
	SourceEnv     Source = "env"
	SourceKeyring Source = "keyring"
	SourceFile    Source = "file"
	SourceAbsent  Source = "absent"
)

// Provider is one supported secret holder. Add an entry here to expose the
// provider through the CLI + TUI + doctor.
type Provider struct {
	Name   string // e.g. "anthropic"
	EnvVar string // e.g. "ANTHROPIC_API_KEY"
	Label  string // human-readable label
}

// Providers returns the known provider set, sorted by Name for stable UIs.
func Providers() []Provider {
	ps := []Provider{
		{Name: "anthropic", EnvVar: "ANTHROPIC_API_KEY", Label: "Anthropic"},
		{Name: "openai", EnvVar: "OPENAI_API_KEY", Label: "OpenAI"},
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].Name < ps[j].Name })
	return ps
}

// ProviderByName looks up a provider by Name. Returns a zero Provider and
// false if no provider matches.
func ProviderByName(name string) (Provider, bool) {
	for _, p := range Providers() {
		if p.Name == name {
			return p, true
		}
	}
	return Provider{}, false
}

// Store is the interface every secret backend satisfies. The default Store
// resolves env > keyring > file.
type Store interface {
	Get(provider string) (string, Source, error)
	Set(provider, value string) (Source, error)
	Delete(provider string) error
}

// NewStore returns the default store. filePath is the path to the fallback
// secrets file; empty means "use the XDG default".
func NewStore(filePath string) (Store, error) {
	if filePath == "" {
		p, err := defaultFilePath()
		if err != nil {
			return nil, err
		}
		filePath = p
	}
	return &layeredStore{filePath: filePath}, nil
}

// DefaultFilePath returns the path at which the file fallback lives.
// Exposed for doctor and TUI display. Never read eagerly - the file may not
// exist until Set is called.
func DefaultFilePath() (string, error) { return defaultFilePath() }

func defaultFilePath() (string, error) {
	if p, err := xdg.ConfigFile("humblskills/secrets.json"); err == nil {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve secrets path: %w", err)
	}
	return filepath.Join(home, ".humblskills", "secrets.json"), nil
}

type layeredStore struct {
	filePath string
}

func (s *layeredStore) Get(provider string) (string, Source, error) {
	p, ok := ProviderByName(provider)
	if !ok {
		return "", SourceAbsent, fmt.Errorf("unknown provider %q", provider)
	}
	if v := strings.TrimSpace(os.Getenv(p.EnvVar)); v != "" {
		return v, SourceEnv, nil
	}
	if v, err := keyring.Get(Service, keyringAccount(provider)); err == nil && v != "" {
		return v, SourceKeyring, nil
	} else if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		// Keyring unavailable (no dbus on Linux, blocked on macOS, etc.).
		// Fall through to file rather than failing.
	}
	if v, ok := readFile(s.filePath, provider); ok {
		return v, SourceFile, nil
	}
	return "", SourceAbsent, nil
}

func (s *layeredStore) Set(provider, value string) (Source, error) {
	if _, ok := ProviderByName(provider); !ok {
		return SourceAbsent, fmt.Errorf("unknown provider %q", provider)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return SourceAbsent, errors.New("refusing to store empty secret")
	}
	if err := keyring.Set(Service, keyringAccount(provider), value); err == nil {
		// Also remove any stale file copy so file > keyring precedence
		// disagreements can't arise.
		_ = writeFileSecret(s.filePath, provider, "")
		return SourceKeyring, nil
	}
	if err := writeFileSecret(s.filePath, provider, value); err != nil {
		return SourceAbsent, fmt.Errorf("keyring unavailable and file fallback failed: %w", err)
	}
	return SourceFile, nil
}

func (s *layeredStore) Delete(provider string) error {
	if _, ok := ProviderByName(provider); !ok {
		return fmt.Errorf("unknown provider %q", provider)
	}
	_ = keyring.Delete(Service, keyringAccount(provider))
	return writeFileSecret(s.filePath, provider, "")
}

// --- file fallback ----------------------------------------------------------

type fileDoc struct {
	SchemaVersion int               `json:"schema_version"`
	Keys          map[string]string `json:"keys"`
}

func readFile(path, provider string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	var d fileDoc
	if err := json.Unmarshal(data, &d); err != nil {
		return "", false
	}
	v, ok := d.Keys[provider]
	if !ok || strings.TrimSpace(v) == "" {
		return "", false
	}
	return v, true
}

func writeFileSecret(path, provider, value string) error {
	var d fileDoc
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &d)
	}
	if d.Keys == nil {
		d.Keys = map[string]string{}
	}
	d.SchemaVersion = 1
	if value == "" {
		delete(d.Keys, provider)
	} else {
		d.Keys[provider] = value
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create secrets dir: %w", err)
	}
	// Remove the file if we emptied it.
	if len(d.Keys) == 0 {
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("remove secrets file: %w", err)
		}
		return nil
	}
	out, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o600); err != nil {
		return fmt.Errorf("write tmp secrets: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename tmp secrets: %w", err)
	}
	return nil
}

func keyringAccount(provider string) string { return provider + "-api-key" }

// --- registry auth token ----------------------------------------------------
//
// The registry token authenticates fetches from a private skill registry. It is
// deliberately NOT an LLM Provider (so it stays out of Providers(), doctor's
// provider list, and the eval key prompts) but reuses the same
// env > keyring > file resolution and the same on-disk secrets file.

// RegistryTokenEnvVar is the environment variable checked first when resolving
// the registry auth token.
const RegistryTokenEnvVar = "HUMBLSKILLS_TOKEN"

// registryTokenKey is the keyring account and secrets-file key for the token.
const registryTokenKey = "registry-token"

// registryAccount is the keyring account / secrets-file key for a registry
// token. The empty name is the default (single-registry) token; a named
// registry gets "registry-token-<name>".
func registryAccount(name string) string {
	if name == "" {
		return registryTokenKey
	}
	return registryTokenKey + "-" + name
}

// GetRegistryToken resolves the default registry token, preferring the
// environment, then the OS keyring, then the 0600 secrets file. Returns
// ("", SourceAbsent) when unset.
func GetRegistryToken() (string, Source) {
	path, _ := defaultFilePath()
	return getRegistryToken(path)
}

func getRegistryToken(filePath string) (string, Source) {
	if v := strings.TrimSpace(os.Getenv(RegistryTokenEnvVar)); v != "" {
		return v, SourceEnv
	}
	return getKeyedToken(filePath, registryAccount(""))
}

// GetRegistryTokenFor resolves a named registry's token: its own keyring/file
// entry first, then the default token (env > keyring > file) as a fallback, so a
// single stored token can cover registries without a dedicated one.
func GetRegistryTokenFor(name string) (string, Source) {
	if name == "" {
		return GetRegistryToken()
	}
	path, _ := defaultFilePath()
	if v, src := getKeyedToken(path, registryAccount(name)); src != SourceAbsent {
		return v, src
	}
	return GetRegistryToken()
}

func getKeyedToken(filePath, account string) (string, Source) {
	if v, err := keyring.Get(Service, account); err == nil && v != "" {
		return v, SourceKeyring
	}
	if filePath != "" {
		if v, ok := readFile(filePath, account); ok {
			return v, SourceFile
		}
	}
	return "", SourceAbsent
}

// SetRegistryToken stores the default registry token, preferring the OS keyring
// and falling back to the 0600 secrets file when the keyring is unavailable.
func SetRegistryToken(value string) (Source, error) {
	path, err := defaultFilePath()
	if err != nil {
		return SourceAbsent, err
	}
	return setRegistryToken(path, value)
}

func setRegistryToken(filePath, value string) (Source, error) {
	return setKeyedToken(filePath, registryAccount(""), value)
}

// SetRegistryTokenFor stores a named registry's token.
func SetRegistryTokenFor(name, value string) (Source, error) {
	path, err := defaultFilePath()
	if err != nil {
		return SourceAbsent, err
	}
	return setKeyedToken(path, registryAccount(name), value)
}

func setKeyedToken(filePath, account, value string) (Source, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return SourceAbsent, errors.New("refusing to store empty token")
	}
	if err := keyring.Set(Service, account, value); err == nil {
		// Drop any stale file copy so file/keyring precedence can't disagree.
		_ = writeFileSecret(filePath, account, "")
		return SourceKeyring, nil
	}
	if err := writeFileSecret(filePath, account, value); err != nil {
		return SourceAbsent, fmt.Errorf("keyring unavailable and file fallback failed: %w", err)
	}
	return SourceFile, nil
}

// DeleteRegistryToken removes the default registry token from keyring + file.
func DeleteRegistryToken() error {
	path, err := defaultFilePath()
	if err != nil {
		return err
	}
	return deleteRegistryToken(path)
}

func deleteRegistryToken(filePath string) error {
	return deleteKeyedToken(filePath, registryAccount(""))
}

// DeleteRegistryTokenFor removes a named registry's token.
func DeleteRegistryTokenFor(name string) error {
	path, err := defaultFilePath()
	if err != nil {
		return err
	}
	return deleteKeyedToken(path, registryAccount(name))
}

func deleteKeyedToken(filePath, account string) error {
	_ = keyring.Delete(Service, account)
	return writeFileSecret(filePath, account, "")
}

// RenameRegistryToken moves a named registry's own stored token from old to
// new. If old has no dedicated token, this is a no-op (nil).
func RenameRegistryToken(old, name string) error {
	if old == name {
		return nil
	}
	path, err := defaultFilePath()
	if err != nil {
		return err
	}
	v, src := getKeyedToken(path, registryAccount(old))
	if src == SourceAbsent || v == "" {
		return nil
	}
	if _, err := setKeyedToken(path, registryAccount(name), v); err != nil {
		return err
	}
	return deleteKeyedToken(path, registryAccount(old))
}

// KeyringAvailable reports whether the OS keyring appears reachable. Used by
// doctor to surface a helpful hint when secrets will land in the file.
func KeyringAvailable() bool {
	// Probe with a non-destructive Get against a reserved account name.
	_, err := keyring.Get(Service, "humblskills-probe")
	if err == nil {
		return true
	}
	return errors.Is(err, keyring.ErrNotFound)
}
