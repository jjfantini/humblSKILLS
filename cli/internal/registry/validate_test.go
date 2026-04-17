package registry

import "testing"

func TestValidateDeps_Clean(t *testing.T) {
	r := &Registry{
		Skills: []Skill{
			{Name: "a", Version: "0.1.0"},
			{Name: "b", Version: "0.2.0", Requires: []string{"a@>=0.1.0"}},
		},
	}
	if issues := ValidateDeps(r); len(issues) != 0 {
		t.Fatalf("expected no issues, got %+v", issues)
	}
}

func TestValidateDeps_Unknown(t *testing.T) {
	r := &Registry{
		Skills: []Skill{
			{Name: "a", Version: "0.1.0", Requires: []string{"ghost"}},
		},
	}
	issues := ValidateDeps(r)
	if len(issues) != 1 || issues[0].Kind != IssueUnknown {
		t.Fatalf("got %+v", issues)
	}
}

func TestValidateDeps_Unsatisfied(t *testing.T) {
	r := &Registry{
		Skills: []Skill{
			{Name: "a", Version: "0.1.0"},
			{Name: "b", Version: "0.2.0", Requires: []string{"a@>=0.5.0"}},
		},
	}
	issues := ValidateDeps(r)
	if len(issues) != 1 || issues[0].Kind != IssueUnsatisfied {
		t.Fatalf("got %+v", issues)
	}
}

func TestValidateDeps_Parse(t *testing.T) {
	r := &Registry{
		Skills: []Skill{
			{Name: "a", Version: "0.1.0", Requires: []string{"a@broken"}},
		},
	}
	issues := ValidateDeps(r)
	if len(issues) == 0 || issues[0].Kind != IssueParse {
		t.Fatalf("got %+v", issues)
	}
}

func TestValidateDeps_Cycle(t *testing.T) {
	r := &Registry{
		Skills: []Skill{
			{Name: "a", Version: "0.1.0", Requires: []string{"b"}},
			{Name: "b", Version: "0.1.0", Requires: []string{"a"}},
		},
	}
	issues := ValidateDeps(r)
	foundCycle := false
	for _, i := range issues {
		if i.Kind == IssueCycle {
			foundCycle = true
		}
	}
	if !foundCycle {
		t.Fatalf("expected cycle issue, got %+v", issues)
	}
}

func TestValidateDeps_Nil(t *testing.T) {
	if got := ValidateDeps(nil); got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}
