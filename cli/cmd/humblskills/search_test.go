package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

func TestSearch_NameMatch_JSON(t *testing.T) {
	s := testutil.NewSandbox(t)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{Name: "foo-helper", Version: "1.0.0", Description: "Help with foo",
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
		{Name: "bar", Version: "1.0.0", Description: "Unrelated",
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
	})

	res := runCLIWithStdoutCapture(t,
		"search", "foo",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\n%s", res.RunErr, res.Err)
	}

	idx := strings.Index(res.Out, "{")
	var out struct {
		Query   string `json:"query"`
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(res.Out[idx:]), &out); err != nil {
		t.Fatalf("parse: %v\n%s", err, res.Out)
	}
	if len(out.Results) != 1 || out.Results[0].Name != "foo-helper" {
		t.Errorf("results = %+v", out.Results)
	}
}

func TestSearch_TagMatch_JSON(t *testing.T) {
	s := testutil.NewSandbox(t)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{Name: "alpha", Version: "1.0.0", Tags: []string{"workflow"},
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
		{Name: "beta", Version: "1.0.0",
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
	})
	res := runCLIWithStdoutCapture(t,
		"search", "workflow",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\n%s", res.RunErr, res.Err)
	}
	idx := strings.Index(res.Out, "{")
	var out struct {
		Results []struct{ Name string } `json:"results"`
	}
	_ = json.Unmarshal([]byte(res.Out[idx:]), &out)
	if len(out.Results) != 1 || out.Results[0].Name != "alpha" {
		t.Errorf("results = %+v", out.Results)
	}
}

func TestSearch_CategoryFilter_JSON(t *testing.T) {
	s := testutil.NewSandbox(t)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{Name: "alpha", Version: "1.0.0", Category: "development",
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
		{Name: "beta", Version: "1.0.0", Category: "writing",
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
	})
	res := runCLIWithStdoutCapture(t,
		"search",
		"--category", "development",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\n%s", res.RunErr, res.Err)
	}
	idx := strings.Index(res.Out, "{")
	var out struct {
		Results []struct{ Name string } `json:"results"`
	}
	if err := json.Unmarshal([]byte(res.Out[idx:]), &out); err != nil {
		t.Fatalf("parse: %v\n%s", err, res.Out)
	}
	if len(out.Results) != 1 || out.Results[0].Name != "alpha" {
		t.Errorf("results = %+v", out.Results)
	}
}

func TestSearch_CategoryFilter_CombinesWithQuery(t *testing.T) {
	s := testutil.NewSandbox(t)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{Name: "commit-helper", Version: "1.0.0", Category: "development",
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
		{Name: "commit-poems", Version: "1.0.0", Category: "writing",
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
	})
	res := runCLIWithStdoutCapture(t,
		"search", "commit",
		"--category", "writing",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\n%s", res.RunErr, res.Err)
	}
	idx := strings.Index(res.Out, "{")
	var out struct {
		Results []struct{ Name string } `json:"results"`
	}
	_ = json.Unmarshal([]byte(res.Out[idx:]), &out)
	if len(out.Results) != 1 || out.Results[0].Name != "commit-poems" {
		t.Errorf("results = %+v", out.Results)
	}
}

func TestSearch_UnknownCategory_Errors(t *testing.T) {
	s := testutil.NewSandbox(t)
	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{Name: "alpha", Version: "1.0.0", Category: "development",
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
	})
	res := runCLIWithStdoutCapture(t,
		"search",
		"--category", "astrology",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--json",
	)
	if res.RunErr == nil {
		t.Fatal("expected error for unknown --category")
	}
}

func TestSearch_NoResults_JSON(t *testing.T) {
	s := testutil.NewSandbox(t)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{Name: "alpha", Version: "1.0.0",
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
	})
	res := runCLIWithStdoutCapture(t,
		"search", "nomatch",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\n%s", res.RunErr, res.Err)
	}
	idx := strings.Index(res.Out, "{")
	var out struct {
		Results []any `json:"results"`
	}
	_ = json.Unmarshal([]byte(res.Out[idx:]), &out)
	if len(out.Results) != 0 {
		t.Errorf("expected zero results, got %+v", out.Results)
	}
}

func TestMatches_NameDescriptionTag(t *testing.T) {
	// Indirect unit test via the command helper matches().
	// The function is tested by making sure different fields all hit.
	// Build fixtures where query only matches one field at a time.
	s := testutil.NewSandbox(t)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{Name: "alpha", Version: "1.0.0", Description: "marker-desc",
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
		{Name: "marker-name", Version: "1.0.0",
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
		{Name: "gamma", Version: "1.0.0", Tags: []string{"marker-tag"},
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
	})

	search := func(q string) int {
		r := runCLIWithStdoutCapture(t, "search", q,
			"--registry", regURL, "--cache-dir", s.CacheDir, "--json")
		if r.RunErr != nil {
			t.Fatalf("search %q: %v", q, r.RunErr)
		}
		idx := strings.Index(r.Out, "{")
		var out struct{ Results []any }
		_ = json.Unmarshal([]byte(r.Out[idx:]), &out)
		return len(out.Results)
	}
	if n := search("marker-name"); n != 1 {
		t.Errorf("name match: got %d", n)
	}
	if n := search("marker-desc"); n != 1 {
		t.Errorf("desc match: got %d", n)
	}
	if n := search("marker-tag"); n != 1 {
		t.Errorf("tag match: got %d", n)
	}
}
