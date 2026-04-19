// Package adapters owns the set of agent platforms humblskills supports.
// It defines the Adapter type, embeds the canonical YAML files, loads them
// at runtime, and provides detect/target helpers used by the CLI and
// registry builder.
package adapters

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

// NameSet returns the set of adapter names suitable for passing to the
// frontmatter validator.
func NameSet(as []Adapter) map[string]struct{} {
	out := make(map[string]struct{}, len(as))
	for _, a := range as {
		out[a.Name] = struct{}{}
	}
	return out
}
