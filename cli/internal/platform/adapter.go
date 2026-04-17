// Package platform loads adapter YAML files that describe how to install a
// skill onto a specific agent platform (Claude Code, Cursor, ...).
//
// Phase 1 only consumes adapter names for frontmatter validation. Detection
// rules and install-target resolution land in later phases.
package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Adapter is the full adapter definition. Phase 1 only reads Name; the other
// fields are parsed to surface YAML errors early and reserved for later phases.
type Adapter struct {
	Name           string            `yaml:"name"`
	Detect         DetectRules       `yaml:"detect"`
	InstallTargets map[string]string `yaml:"install_targets"`
	DefaultScope   string            `yaml:"default_scope"`
	Transform      string            `yaml:"transform"`
}

// DetectRules describes how to decide whether this platform is present on the
// user's machine. Only one of AnyOf / AllOf should be set.
type DetectRules struct {
	AnyOf []DetectRule `yaml:"any_of,omitempty"`
	AllOf []DetectRule `yaml:"all_of,omitempty"`
}

// DetectRule is a single probe. Exactly one field should be set.
type DetectRule struct {
	PathExists string `yaml:"path_exists,omitempty"`
	Env        string `yaml:"env,omitempty"`
}

// LoadAll reads every *.yaml file in dir as an adapter. Adapters are returned
// sorted by name for deterministic output.
func LoadAll(dir string) ([]Adapter, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read adapters dir: %w", err)
	}
	var out []Adapter
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		var a Adapter
		if err := yaml.Unmarshal(data, &a); err != nil {
			return nil, fmt.Errorf("%s: yaml parse: %w", name, err)
		}
		if a.Name == "" {
			return nil, fmt.Errorf("%s: missing `name`", name)
		}
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// NameSet returns the set of adapter names suitable for passing to the
// frontmatter validator.
func NameSet(as []Adapter) map[string]struct{} {
	out := make(map[string]struct{}, len(as))
	for _, a := range as {
		out[a.Name] = struct{}{}
	}
	return out
}
