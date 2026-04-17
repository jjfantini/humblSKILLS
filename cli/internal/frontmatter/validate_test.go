package frontmatter

import (
	"strings"
	"testing"
)

func ctxWith(skills map[string]string, adapters ...string) ValidationContext {
	ads := make(map[string]struct{}, len(adapters))
	for _, a := range adapters {
		ads[a] = struct{}{}
	}
	return ValidationContext{KnownSkills: skills, KnownAdapters: ads}
}

func TestValidate_Happy(t *testing.T) {
	fm := Frontmatter{
		Name:        "foo",
		Description: "desc",
		Version:     "0.1.0",
		Platforms:   []string{"claude-code"},
	}
	if err := fm.Validate("foo", ctxWith(nil, "claude-code")); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestValidate_NameRegex(t *testing.T) {
	fm := Frontmatter{Name: "Foo_Bar", Description: "d", Version: "0.1.0"}
	err := fm.Validate("Foo_Bar", ctxWith(nil))
	if err == nil || !strings.Contains(err.Error(), "must match") {
		t.Fatalf("expected name regex error, got %v", err)
	}
}

func TestValidate_NameMatchesDir(t *testing.T) {
	fm := Frontmatter{Name: "foo", Description: "d", Version: "0.1.0"}
	err := fm.Validate("bar", ctxWith(nil))
	if err == nil || !strings.Contains(err.Error(), "containing directory") {
		t.Fatalf("expected dir-mismatch error, got %v", err)
	}
}

func TestValidate_SemverBad(t *testing.T) {
	fm := Frontmatter{Name: "foo", Description: "d", Version: "1.2"}
	err := fm.Validate("foo", ctxWith(nil))
	if err == nil || !strings.Contains(err.Error(), "not valid semver") {
		t.Fatalf("expected semver error, got %v", err)
	}
}

func TestValidate_DepUnknown(t *testing.T) {
	fm := Frontmatter{Name: "foo", Description: "d", Version: "0.1.0", Requires: []string{"ghost"}}
	err := fm.Validate("foo", ctxWith(nil))
	if err == nil || !strings.Contains(err.Error(), `unknown dep "ghost"`) {
		t.Fatalf("expected unknown dep error, got %v", err)
	}
}

func TestValidate_DepVersionUnsatisfied(t *testing.T) {
	fm := Frontmatter{
		Name:        "foo",
		Description: "d",
		Version:     "0.1.0",
		Requires:    []string{"bar@>=0.3.0"},
	}
	err := fm.Validate("foo", ctxWith(map[string]string{"bar": "0.2.0"}))
	if err == nil || !strings.Contains(err.Error(), "unsatisfied") {
		t.Fatalf("expected unsatisfied error, got %v", err)
	}
}

func TestValidate_DepVersionSatisfied(t *testing.T) {
	fm := Frontmatter{
		Name:        "foo",
		Description: "d",
		Version:     "0.1.0",
		Requires:    []string{"bar@>=0.3.0"},
	}
	err := fm.Validate("foo", ctxWith(map[string]string{"bar": "0.3.1"}))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestValidate_UnknownPlatform(t *testing.T) {
	fm := Frontmatter{
		Name: "foo", Description: "d", Version: "0.1.0",
		Platforms: []string{"claude-code", "atari-2600"},
	}
	err := fm.Validate("foo", ctxWith(nil, "claude-code"))
	if err == nil || !strings.Contains(err.Error(), "atari-2600") {
		t.Fatalf("expected unknown platform error, got %v", err)
	}
}

func TestValidate_SelfDep(t *testing.T) {
	fm := Frontmatter{
		Name: "foo", Description: "d", Version: "0.1.0",
		Requires: []string{"foo"},
	}
	err := fm.Validate("foo", ctxWith(map[string]string{"foo": "0.1.0"}))
	if err == nil || !strings.Contains(err.Error(), "cannot require itself") {
		t.Fatalf("expected self-dep error, got %v", err)
	}
}

func TestParseDep_Forms(t *testing.T) {
	cases := []struct {
		in       string
		wantName string
		wantOp   string
		wantVer  string
		wantErr  bool
	}{
		{"foo", "foo", "", "", false},
		{"foo@1.2.3", "foo", "==", "1.2.3", false},
		{"foo@>=0.2.0", "foo", ">=", "0.2.0", false},
		{"", "", "", "", true},
		{"@1.0.0", "", "", "", true},
		{"foo@", "", "", "", true},
		{"foo@notsemver", "", "", "", true},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			d, err := ParseDep(c.in)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected: %v", err)
			}
			if d.Name != c.wantName || d.Op != c.wantOp || d.Version != c.wantVer {
				t.Errorf("got %+v, want name=%q op=%q ver=%q", d, c.wantName, c.wantOp, c.wantVer)
			}
		})
	}
}

func TestDepSatisfies(t *testing.T) {
	cases := []struct {
		dep        string
		registered string
		want       bool
	}{
		{"foo", "0.1.0", true},
		{"foo@0.1.0", "0.1.0", true},
		{"foo@0.1.0", "0.1.1", false},
		{"foo@>=0.1.0", "0.1.0", true},
		{"foo@>=0.1.0", "0.0.9", false},
		{"foo@>=0.1.0", "1.0.0", true},
	}
	for _, c := range cases {
		t.Run(c.dep+"/"+c.registered, func(t *testing.T) {
			d, err := ParseDep(c.dep)
			if err != nil {
				t.Fatal(err)
			}
			if got := d.Satisfies(c.registered); got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}
