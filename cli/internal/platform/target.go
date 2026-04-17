package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Target is a resolved install target for one adapter/scope combination.
type Target struct {
	Scope    string
	Path     string // expanded, absolute when possible
	Writable bool
}

// Targets returns every declared install target for this adapter, in scope-
// sorted order ("project", "user" lexical). Paths are expanded and probed
// for writability.
func (a Adapter) Targets() []Target {
	scopes := make([]string, 0, len(a.InstallTargets))
	for s := range a.InstallTargets {
		scopes = append(scopes, s)
	}
	sort.Strings(scopes)

	out := make([]Target, 0, len(scopes))
	for _, s := range scopes {
		raw := a.InstallTargets[s]
		p := ExpandPath(raw)
		if !filepath.IsAbs(p) {
			// Project-scoped paths are relative to CWD; make absolute for
			// display. If CWD lookup fails, leave the relative form.
			if cwd, err := os.Getwd(); err == nil {
				p = filepath.Join(cwd, p)
			}
		}
		out = append(out, Target{
			Scope:    s,
			Path:     filepath.Clean(p),
			Writable: isWritable(p),
		})
	}
	return out
}

// Target returns the install target for the requested scope. If scope is
// empty, DefaultScope is used. Returns an error if the scope isn't declared
// for this adapter.
func (a Adapter) Target(scope string) (Target, error) {
	if scope == "" {
		scope = a.DefaultScope
	}
	raw, ok := a.InstallTargets[scope]
	if !ok {
		return Target{}, fmt.Errorf("adapter %q has no install target for scope %q", a.Name, scope)
	}
	p := ExpandPath(raw)
	if !filepath.IsAbs(p) {
		if cwd, err := os.Getwd(); err == nil {
			p = filepath.Join(cwd, p)
		}
	}
	p = filepath.Clean(p)
	return Target{Scope: scope, Path: p, Writable: isWritable(p)}, nil
}

// isWritable reports whether the CLI can create files at path. If path doesn't
// exist, it walks up to the first existing parent and probes that.
func isWritable(path string) bool {
	dir := path
	for {
		info, err := os.Stat(dir)
		if err == nil {
			if info.IsDir() {
				break
			}
			return false
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return false
		}
		dir = parent
	}
	f, err := os.CreateTemp(dir, ".humblskills-writable-*")
	if err != nil {
		return false
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return true
}
