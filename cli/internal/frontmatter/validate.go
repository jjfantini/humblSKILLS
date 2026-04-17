package frontmatter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/mod/semver"
)

// NameRegex is the allowed shape for a skill name.
var NameRegex = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// isStrictSemver accepts only MAJOR.MINOR.PATCH, optionally with prerelease
// or build metadata. Go's x/mod/semver treats "1.2" as valid (v1.2.0); we
// want to reject shortened forms.
func isStrictSemver(v string) bool {
	if !semver.IsValid("v" + v) {
		return false
	}
	noMeta := v
	if i := strings.IndexAny(noMeta, "-+"); i >= 0 {
		noMeta = noMeta[:i]
	}
	return strings.Count(noMeta, ".") == 2
}

// Dep is a parsed entry from the `requires:` list.
type Dep struct {
	Name    string
	Op      string // "", ">=", "=="
	Version string // semver, empty when Op is ""
}

// ParseDep parses a single requires entry.
//
// Supported syntaxes:
//
//	"foo"            -> any version
//	"foo@1.2.3"      -> exact match
//	"foo@>=1.2.3"    -> minimum version
func ParseDep(s string) (Dep, error) {
	if s == "" {
		return Dep{}, errors.New("empty dep")
	}
	at := strings.Index(s, "@")
	if at < 0 {
		return Dep{Name: s}, nil
	}
	name := s[:at]
	spec := s[at+1:]
	if name == "" {
		return Dep{}, errors.New("dep name is empty")
	}
	if spec == "" {
		return Dep{}, errors.New("dep version spec is empty after '@'")
	}

	var op, ver string
	switch {
	case strings.HasPrefix(spec, ">="):
		op = ">="
		ver = strings.TrimSpace(spec[2:])
	default:
		op = "=="
		ver = spec
	}
	if !isStrictSemver(ver) {
		return Dep{}, fmt.Errorf("version %q is not valid semver (want MAJOR.MINOR.PATCH)", ver)
	}
	return Dep{Name: name, Op: op, Version: ver}, nil
}

// Satisfies reports whether the given registered version meets this dep's
// constraint. An empty Op always satisfies.
func (d Dep) Satisfies(registered string) bool {
	if d.Op == "" {
		return true
	}
	if !semver.IsValid("v" + registered) {
		return false
	}
	cmp := semver.Compare("v"+registered, "v"+d.Version)
	switch d.Op {
	case "==":
		return cmp == 0
	case ">=":
		return cmp >= 0
	default:
		return false
	}
}

// ValidationContext gives Validate access to cross-skill facts that a single
// SKILL.md can't know on its own.
type ValidationContext struct {
	// KnownSkills maps skill name -> registered version.
	KnownSkills map[string]string
	// KnownAdapters is the set of adapter names declared under adapters/.
	KnownAdapters map[string]struct{}
}

// Validate checks every humblSKILLS-owned rule except DAG acyclicity (done by
// the resolver). dirName is the skill directory's basename; skillPath is the
// absolute path to that directory, used only to stat post_install.
func (fm Frontmatter) Validate(dirName, skillPath string, ctx ValidationContext) error {
	var errs []string

	switch {
	case fm.Name == "":
		errs = append(errs, "name is required")
	default:
		if !NameRegex.MatchString(fm.Name) {
			errs = append(errs, fmt.Sprintf("name %q must match %s", fm.Name, NameRegex))
		}
		if fm.Name != dirName {
			errs = append(errs, fmt.Sprintf("name %q must match containing directory %q", fm.Name, dirName))
		}
	}

	if strings.TrimSpace(fm.Description) == "" {
		errs = append(errs, "description is required")
	}

	switch {
	case fm.Version == "":
		errs = append(errs, "version is required")
	case !isStrictSemver(fm.Version):
		errs = append(errs, fmt.Sprintf("version %q is not valid semver (want MAJOR.MINOR.PATCH)", fm.Version))
	}

	for _, plat := range fm.Platforms {
		if _, ok := ctx.KnownAdapters[plat]; !ok {
			errs = append(errs, fmt.Sprintf("unknown platform %q (no matching adapter)", plat))
		}
	}

	for _, raw := range fm.Requires {
		dep, err := ParseDep(raw)
		if err != nil {
			errs = append(errs, fmt.Sprintf("invalid dep %q: %v", raw, err))
			continue
		}
		if dep.Name == fm.Name {
			errs = append(errs, fmt.Sprintf("skill cannot require itself (%q)", raw))
			continue
		}
		registered, ok := ctx.KnownSkills[dep.Name]
		if !ok {
			errs = append(errs, fmt.Sprintf("unknown dep %q", dep.Name))
			continue
		}
		if !dep.Satisfies(registered) {
			errs = append(errs, fmt.Sprintf("dep %q unsatisfied by registered version %s", raw, registered))
		}
	}

	if fm.PostInstall != "" {
		clean := filepath.Clean(fm.PostInstall)
		if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") || clean == ".." {
			errs = append(errs, fmt.Sprintf("post_install %q must be a relative path inside the skill directory", fm.PostInstall))
		} else {
			full := filepath.Join(skillPath, clean)
			if _, err := os.Stat(full); err != nil {
				errs = append(errs, fmt.Sprintf("post_install %q not found: %v", fm.PostInstall, err))
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}
	label := fm.Name
	if label == "" {
		label = dirName
	}
	return fmt.Errorf("%s: %s", label, strings.Join(errs, "; "))
}
