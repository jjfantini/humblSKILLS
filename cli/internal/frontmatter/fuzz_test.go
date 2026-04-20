package frontmatter_test

import (
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/frontmatter"
)

// FuzzParse ensures the frontmatter parser never panics on arbitrary
// input. Correctness (valid vs invalid) is decided by Parse's returned
// error; the fuzz target only guards the panic contract.
func FuzzParse(f *testing.F) {
	// Seed with examples covering the main branches.
	seeds := []string{
		"",
		"---\n---\n\n# body\n",
		"---\nname: foo\nversion: 1.0.0\ndescription: desc\n---\n",
		"--- not frontmatter",
		"\n\n\n",
		"---\nname: x\n",   // unterminated
		"---\n\t\t\t---\n", // tabs in YAML (invalid)
		"---\nname: [array, not, scalar]\n---\n",
		"---\nversion: 1.2\n---\n", // non-semver
		"---\nrequires:\n  - foo@1.0.0\n  - bar\n---\n",
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _, _ = frontmatter.Parse(data)
	})
}

// FuzzParseDep ensures version-constraint parsing never panics.
func FuzzParseDep(f *testing.F) {
	seeds := []string{
		"",
		"foo",
		"foo@1.0.0",
		"foo@>=1.0.0",
		"foo@~1.0.0",
		"@malformed",
		"foo@",
		"foo@1",
		"foo@1.2",
		"foo@abc",
		"foo@>=", // incomplete constraint
		"@@@",
		"with spaces@1.0.0",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		dep, err := frontmatter.ParseDep(s)
		if err == nil {
			// If parse succeeded, Satisfies must not panic either.
			_ = dep.Satisfies("1.0.0")
			_ = dep.Satisfies("malformed")
		}
	})
}
