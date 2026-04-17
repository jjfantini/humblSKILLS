package platform

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed builtin/*.yaml
var builtinFS embed.FS

// LoadBuiltin returns the adapters baked into the CLI binary. It mirrors the
// canonical adapters/ directory at the repo root; `make sync-adapters` keeps
// the two copies in sync and CI fails on drift.
func LoadBuiltin() ([]Adapter, error) {
	entries, err := fs.ReadDir(builtinFS, "builtin")
	if err != nil {
		return nil, fmt.Errorf("read embedded adapters: %w", err)
	}
	var out []Adapter
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		data, err := fs.ReadFile(builtinFS, "builtin/"+name)
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
