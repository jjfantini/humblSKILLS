package env

import (
	"os"
	"path/filepath"
	"testing"
)

// FuzzParseDotEnv ensures the .env parser never panics on arbitrary
// content. The parser is intentionally lenient — unknown lines are
// skipped rather than errored — so we only guard the panic contract.
func FuzzParseDotEnv(f *testing.F) {
	seeds := []string{
		"",
		"FOO=bar\n",
		"export FOO=\"quoted value\"\n",
		"FOO='single'\n",
		"# comment only\n",
		"FOO\n",
		"=empty-key\n",
		"SPACES =  val  \n",
		"FOO=\n",
		"FOO=bar\nBAR=baz\n",
		"weird line with no equals",
		"\x00\x01\x02",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, body string) {
		dir := t.TempDir()
		p := filepath.Join(dir, ".env")
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Skip()
		}
		_, _ = parse(p)
	})
}
