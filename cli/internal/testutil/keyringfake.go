package testutil

import (
	"testing"

	"github.com/zalando/go-keyring"
)

// UseFakeKeyring installs a process-wide in-memory keyring mock for
// the duration of t. After the test returns, the mock is reinstalled
// (fresh, empty) so leaked state never bleeds across tests.
//
// The real OS keyring is never touched. Safe to use on CI.
func UseFakeKeyring(t testing.TB) {
	t.Helper()
	keyring.MockInit()
	t.Cleanup(func() { keyring.MockInit() })
}

// UseUnavailableKeyring simulates an OS keyring that always fails.
// Used to exercise the secrets package's file-fallback path.
func UseUnavailableKeyring(t testing.TB, err error) {
	t.Helper()
	keyring.MockInitWithError(err)
	t.Cleanup(func() { keyring.MockInit() })
}
