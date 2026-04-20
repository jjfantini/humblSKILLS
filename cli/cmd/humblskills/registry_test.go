package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

func TestRegistryRefresh_FileURLReportsFileOrigin(t *testing.T) {
	s := testutil.NewSandbox(t)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{Name: "foo", Version: "1.0.0",
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
	})

	res := runCLIWithStdoutCapture(t,
		"registry", "refresh",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\n%s", res.RunErr, res.Err)
	}
	idx := strings.Index(res.Out, "{")
	if idx < 0 {
		t.Fatalf("no JSON in output:\n%s", res.Out)
	}
	var out struct {
		URL    string `json:"url"`
		Source string `json:"source"`
		Skills int    `json:"skills"`
	}
	if err := json.Unmarshal([]byte(res.Out[idx:]), &out); err != nil {
		t.Fatalf("parse: %v\n%s", err, res.Out)
	}
	if out.Source != "file" {
		t.Errorf("source = %q, want file", out.Source)
	}
	if out.Skills != 1 {
		t.Errorf("skills = %d", out.Skills)
	}
}
