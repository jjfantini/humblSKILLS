package adapters_test

import (
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
)

// FuzzExpandPath ensures path expansion never panics on arbitrary
// input — notably on strings containing null bytes, invalid UTF-8,
// path traversal sequences, or nested $VAR references.
func FuzzExpandPath(f *testing.F) {
	seeds := []string{
		"",
		"~",
		"~/foo",
		"$HOME/bar",
		"${HOME}/baz",
		"/absolute/path",
		"../escape",
		"/$HOME/nested/$USER",
		"~$SHELL", // weird: not a tilde-home pattern
		"\x00null/bytes",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		_ = adapters.ExpandPath(s)
	})
}
