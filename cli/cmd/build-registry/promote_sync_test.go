package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/v2/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/frontmatter"
)

// promoteEngine is smart-skill's promotion gate. It re-declares the closed
// sets this command validates against because it runs on machines that only
// have the humblskills binary, so nothing else keeps the two copies equal.
var promoteEngine = filepath.Join("..", "..", "..", "skills", "smart-skill", "scripts", "lib", "promote.py")

// TestPromoteEngineMirrorsClosedSets fails when a category, role, or adapter
// is added here but not in promote.py (or the reverse). A drift either lets
// promote.sh open a PR this build then rejects, or blocks a skill it accepts.
func TestPromoteEngineMirrorsClosedSets(t *testing.T) {
	src, err := os.ReadFile(promoteEngine)
	if err != nil {
		t.Fatalf("read %s: %v", promoteEngine, err)
	}
	adapterList, err := adapters.Load()
	if err != nil {
		t.Fatalf("load adapters: %v", err)
	}
	platforms := make([]string, 0, len(adapterList))
	for _, a := range adapterList {
		platforms = append(platforms, a.Name)
	}

	for _, tc := range []struct {
		pyName string
		goSet  []string
		source string
	}{
		{"CATEGORIES", frontmatter.Categories, "frontmatter.Categories"},
		{"ROLES", frontmatter.Roles, "frontmatter.Roles"},
		{"PLATFORMS", platforms, "the embedded adapters"},
	} {
		got := pythonTuple(t, string(src), tc.pyName)
		if !sameSet(got, tc.goSet) {
			t.Errorf("promote.py %s = %v, but %s = %v; update the tuple in %s",
				tc.pyName, got, tc.source, tc.goSet, promoteEngine)
		}
	}
}

// pythonTuple extracts the quoted members of a module-level `NAME = (...)`
// tuple assignment.
func pythonTuple(t *testing.T, src, name string) []string {
	t.Helper()
	m := regexp.MustCompile(`(?m)^` + name + ` = \(([^)]*)\)`).FindStringSubmatch(src)
	if m == nil {
		t.Fatalf("promote.py has no `%s = (...)` tuple", name)
	}
	var out []string
	for _, q := range regexp.MustCompile(`"([^"]*)"`).FindAllStringSubmatch(m[1], -1) {
		out = append(out, q[1])
	}
	return out
}

func sameSet(a, b []string) bool {
	x := append([]string(nil), a...)
	y := append([]string(nil), b...)
	sort.Strings(x)
	sort.Strings(y)
	return strings.Join(x, "\x00") == strings.Join(y, "\x00")
}
