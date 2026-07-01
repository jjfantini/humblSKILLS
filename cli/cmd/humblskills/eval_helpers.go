package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/scenarios"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/workspace"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/secrets"
)

// This file holds the shared, side-effect-light helpers for the eval command
// family. Command wiring lives in eval.go and the run logic in eval_actions.go.

func resolveSkill(app *App, skill string) (skillDir string, f *scenarios.File, err error) {
	if skill == "" {
		return "", nil, errors.New("skill name required")
	}
	// Collect every candidate location. Order is a preference: a local
	// dev copy with evals/ wins over an installed copy without one, so
	// authoring scenarios against the repo checkout "just works".
	var candidates []string
	if root := os.Getenv("HUMBLSKILLS_ROOT"); root != "" {
		candidates = append(candidates,
			filepath.Join(root, "skills", skill),
			filepath.Join(root, skill),
		)
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "skills", skill),
			filepath.Join(cwd, skill),
		)
		// Running `go run ./cmd/humblskills` from the cli/ module: repo skills live in ../skills/.
		if filepath.Base(cwd) == "cli" {
			candidates = append(candidates, filepath.Join(filepath.Dir(cwd), "skills", skill))
		}
	}
	if m, err := manifest.Load(app.Config.ManifestPath); err == nil && m != nil {
		for _, inst := range m.FindAll(skill) {
			candidates = append(candidates, inst.Path)
		}
	}
	// Prefer candidates whose scenarios.json parses cleanly.
	var firstExisting string
	for _, c := range candidates {
		if _, err := os.Stat(c); err != nil {
			continue
		}
		if firstExisting == "" {
			firstExisting = c
		}
		sf, serr := scenarios.LoadFromSkill(c)
		if serr == nil {
			return c, sf, nil
		}
		if app.Config.Verbose {
			app.UI.Warn("skip %s: %v", c, serr)
		}
	}
	if firstExisting == "" {
		return "", nil, fmt.Errorf("skill %q not found in manifest and no local copy at ./skills/%s", skill, skill)
	}
	// Found the skill but no valid evals — return dir + nil file so the
	// caller can `eval init` into it.
	return firstExisting, nil, nil
}

func resolveWorkspace(app *App, override string) string {
	r := workspace.Resolver{
		FlagOverride: override,
		EnvOverride:  os.Getenv("HUMBLSKILLS_EVAL_WORKSPACE"),
	}
	if p, err := profile.Load(app.Config.ProfilePath); err == nil && p != nil && p.Eval != nil {
		r.ProfileDefault = p.Eval.DefaultWorkspace
	}
	root, err := r.Root()
	if err != nil {
		root, _ = workspace.DefaultRoot()
	}
	return root
}

func pickRunner(reg *runner.Registry, name string) (runner.Runner, error) {
	// Prefer flag, then env, then profile, then auto-detect.
	if name == "" {
		name = os.Getenv("HUMBLSKILLS_EVAL_RUNNER")
	}
	if name == "" {
		return reg.AutoPick(context.Background())
	}
	return reg.ByName(name)
}

func openInBrowser(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

func loadJSON[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func skillBasename(dir string) string { return filepath.Base(dir) }

func scaffoldEvalsDir(dir, skillName string) error {
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("%s already exists — delete it first to re-scaffold", dir)
	}
	if err := os.MkdirAll(filepath.Join(dir, "files"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "assertions"), 0o755); err != nil {
		return err
	}
	scenariosBody := fmt.Sprintf(`{
  "skill_name": %q,
  "schema_version": 1,
  "configurations": ["smart_skill", "flat_skill", "no_skill"],
  "runs_per_configuration": 1,
  "scenarios": [
    {
      "id": "starter",
      "family": "generic",
      "sessions": [
        {
          "n": 1,
          "prompt": "Describe the first task you want %s to handle.",
          "assertions": [
            {"text": "agent produced at least one output file", "check": "path_exists:mock-output.txt"}
          ]
        }
      ]
    }
  ]
}
`, skillName, skillName)
	evalsBody := fmt.Sprintf(`{
  "skill_name": %q,
  "evals": [
    {"id": 1, "prompt": "single-session eval compatible with agentskills.io",
     "expected_output": "describe the expected output",
     "assertions": [{"text": "agent produced output"}]}
  ]
}
`, skillName)
	readme := fmt.Sprintf(`# evals/ for %s

scenarios.json    — humblSKILLS longitudinal + multi-arm scenarios
evals.json        — agentskills.io-compatible single-session evals (optional)
files/            — input fixtures referenced by prompts
assertions/       — optional shell/python scripts for deterministic checks

Run:  humblskills eval run %s
`, skillName, skillName)
	files := map[string]string{
		"scenarios.json": scenariosBody,
		"evals.json":     evalsBody,
		"README.md":      readme,
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func sortedStrKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func unionKeys[V any](a, b map[string]V) []string {
	set := map[string]struct{}{}
	for k := range a {
		set[k] = struct{}{}
	}
	for k := range b {
		set[k] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func knownProviders() []string {
	ps := secrets.Providers()
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.Name)
	}
	return out
}

func abbreviate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}
