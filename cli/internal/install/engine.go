package install

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/fetch"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/platform"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
)

// Outcome classifies what happened for one (skill, platform, scope) triple.
type Outcome string

const (
	// OutcomeInstalled means the target did not previously exist.
	OutcomeInstalled Outcome = "installed"
	// OutcomeReplaced means the target existed but content or source_sha drifted.
	OutcomeReplaced Outcome = "replaced"
	// OutcomeSkipped means the target is already up-to-date.
	OutcomeSkipped Outcome = "skipped"
	// OutcomeForced means --force replaced an up-to-date target anyway.
	OutcomeForced Outcome = "forced"
)

// TargetResult is the outcome of installing one skill onto one adapter+scope.
type TargetResult struct {
	Skill    string  `json:"skill"`
	Platform string  `json:"platform"`
	Scope    string  `json:"scope"`
	Path     string  `json:"path"`
	Outcome  Outcome `json:"outcome"`
}

// Result aggregates every target outcome produced by one install run.
type Result struct {
	Results []TargetResult `json:"results"`
}

// Engine installs skills by fetching tarballs, verifying dir_sha, and copying
// into each requested target.
type Engine struct {
	Fetcher      *fetch.Fetcher
	ManifestPath string
	StagingDir   string
	Now          func() time.Time
}

// NewEngine returns an Engine configured with sane defaults. cacheDir is used
// both for tarball caching and for staging extracted skill trees.
func NewEngine(cacheDir, manifestPath string) *Engine {
	return &Engine{
		Fetcher:      fetch.NewFetcher(cacheDir),
		ManifestPath: manifestPath,
		StagingDir:   filepath.Join(cacheDir, "staging"),
		Now:          time.Now,
	}
}

// ExecuteOpts configures one Execute call.
type ExecuteOpts struct {
	// Adapters is the full set of runtime adapters; only those in Platforms
	// will receive installs.
	Adapters []platform.Adapter
	// Platforms are the adapter names to install onto (order preserved).
	Platforms []string
	// Scope selects the install target scope for every adapter. Empty means
	// the adapter's DefaultScope.
	Scope string
	// Force causes existing targets to be replaced even when the manifest
	// shows they're up-to-date.
	Force bool
}

// Execute runs the plan: fetch each skill, verify its content hash, and place
// it into every requested target, updating the manifest in place.
func (e *Engine) Execute(reg *registry.Registry, plan []Step, opts ExecuteOpts) (Result, error) {
	var res Result
	if reg == nil {
		return res, fmt.Errorf("execute: nil registry")
	}
	if len(plan) == 0 {
		return res, nil
	}

	adapterIndex := make(map[string]platform.Adapter, len(opts.Adapters))
	for _, a := range opts.Adapters {
		adapterIndex[a.Name] = a
	}
	for _, name := range opts.Platforms {
		if _, ok := adapterIndex[name]; !ok {
			return res, fmt.Errorf("unknown adapter %q", name)
		}
	}

	m, err := manifest.Load(e.ManifestPath)
	if err != nil {
		return res, fmt.Errorf("load manifest: %w", err)
	}

	for _, step := range plan {
		stepRes, err := e.installOne(reg, step, opts, adapterIndex, m)
		if err != nil {
			return res, fmt.Errorf("%s: %w", step.Skill.Name, err)
		}
		res.Results = append(res.Results, stepRes...)
	}

	if err := manifest.Save(e.ManifestPath, m); err != nil {
		return res, fmt.Errorf("save manifest: %w", err)
	}
	return res, nil
}

func (e *Engine) installOne(
	reg *registry.Registry,
	step Step,
	opts ExecuteOpts,
	adapterIndex map[string]platform.Adapter,
	m *manifest.Manifest,
) ([]TargetResult, error) {
	skill := step.Skill

	// Decide which targets this skill needs. A skill's platforms[] list acts
	// as an opt-in whitelist; only adapters in both requested-platforms and
	// skill.Platforms get an install.
	allow := map[string]struct{}{}
	if len(skill.Platforms) == 0 {
		for _, p := range opts.Platforms {
			allow[p] = struct{}{}
		}
	} else {
		for _, p := range skill.Platforms {
			allow[p] = struct{}{}
		}
	}

	type pending struct {
		adapter platform.Adapter
		target  platform.Target
	}
	var pendings []pending
	for _, p := range opts.Platforms {
		if _, ok := allow[p]; !ok {
			continue
		}
		adapter := adapterIndex[p]
		scope := opts.Scope
		if scope == "" {
			scope = adapter.DefaultScope
		}
		t, err := adapter.Target(scope)
		if err != nil {
			return nil, err
		}
		pendings = append(pendings, pending{adapter: adapter, target: t})
	}
	if len(pendings) == 0 {
		return nil, nil
	}

	// Pre-compute which targets actually need the extract + copy. If none do,
	// we skip the tarball fetch entirely so repeated no-op installs are fast.
	type planTarget struct {
		pending pending
		out     Outcome
		final   string
		orphan  string // previous install path to clean up (project-scope move)
	}
	var toWrite []planTarget
	var skipped []TargetResult
	for _, pg := range pendings {
		dest := filepath.Join(pg.target.Path, skill.Name)
		existing := m.FindOne(skill.Name, pg.adapter.Name, pg.target.Scope)
		// Project-scope installs are pinned to the CWD that installed them.
		// When the manifest entry's path differs from the current resolved
		// target, treat it as a move: clean up the old location before we
		// write the new one.
		orphan := ""
		if existing != nil && existing.Path != dest {
			orphan = existing.Path
		}

		upToDate := existing != nil &&
			existing.Version == skill.Version &&
			existing.SourceSHA == reg.Source.SHA &&
			existing.RegistryRef == skill.DirSHA &&
			existing.Path == dest
		if _, err := os.Stat(dest); err != nil {
			upToDate = false
		}

		switch {
		case upToDate && !opts.Force:
			skipped = append(skipped, TargetResult{
				Skill: skill.Name, Platform: pg.adapter.Name, Scope: pg.target.Scope,
				Path: dest, Outcome: OutcomeSkipped,
			})
		case upToDate && opts.Force:
			toWrite = append(toWrite, planTarget{pending: pg, out: OutcomeForced, final: dest, orphan: orphan})
		case existing != nil:
			toWrite = append(toWrite, planTarget{pending: pg, out: OutcomeReplaced, final: dest, orphan: orphan})
		default:
			toWrite = append(toWrite, planTarget{pending: pg, out: OutcomeInstalled, final: dest, orphan: orphan})
		}
	}

	if len(toWrite) == 0 {
		return skipped, nil
	}

	staging := filepath.Join(e.StagingDir, skill.Name+"-"+shortSHA(reg.Source.SHA))
	if err := os.RemoveAll(staging); err != nil {
		return nil, fmt.Errorf("clean staging: %w", err)
	}

	tarPath, err := e.Fetcher.Fetch(reg.Source.Repo, reg.Source.SHA)
	if err != nil {
		return nil, err
	}
	if err := fetch.Extract(tarPath, skill.Path, staging); err != nil {
		return nil, fmt.Errorf("extract: %w", err)
	}

	gotSHA, err := registry.DirSHA(staging)
	if err != nil {
		return nil, fmt.Errorf("hash staging: %w", err)
	}
	if gotSHA != skill.DirSHA {
		return nil, fmt.Errorf("dir_sha mismatch for %s: want %s got %s", skill.Name, skill.DirSHA, gotSHA)
	}

	results := append([]TargetResult(nil), skipped...)
	for _, pt := range toWrite {
		if pt.orphan != "" && pt.orphan != pt.final {
			if err := os.RemoveAll(pt.orphan); err != nil {
				return nil, fmt.Errorf("clean old install %s: %w", pt.orphan, err)
			}
		}
		if err := replaceDir(staging, pt.final); err != nil {
			return nil, fmt.Errorf("place %s: %w", pt.final, err)
		}
		m.Upsert(manifest.Installation{
			Skill:       skill.Name,
			Version:     skill.Version,
			Platform:    pt.pending.adapter.Name,
			Scope:       pt.pending.target.Scope,
			Path:        pt.final,
			InstalledAt: e.Now().UTC(),
			SourceSHA:   reg.Source.SHA,
			RegistryRef: skill.DirSHA,
		})
		results = append(results, TargetResult{
			Skill:    skill.Name,
			Platform: pt.pending.adapter.Name,
			Scope:    pt.pending.target.Scope,
			Path:     pt.final,
			Outcome:  pt.out,
		})
	}
	return results, nil
}

// Uninstall removes every on-disk target for skill and deletes its manifest
// entries. Missing directories are tolerated — the manifest is still cleaned
// up so subsequent installs are idempotent.
func (e *Engine) Uninstall(skill string) ([]TargetResult, error) {
	m, err := manifest.Load(e.ManifestPath)
	if err != nil {
		return nil, fmt.Errorf("load manifest: %w", err)
	}
	entries := m.FindAll(skill)
	if len(entries) == 0 {
		return nil, nil
	}
	var results []TargetResult
	for _, e := range entries {
		if err := os.RemoveAll(e.Path); err != nil {
			return results, fmt.Errorf("remove %s: %w", e.Path, err)
		}
		results = append(results, TargetResult{
			Skill: e.Skill, Platform: e.Platform, Scope: e.Scope,
			Path: e.Path, Outcome: "removed",
		})
	}
	m.Remove(skill)
	if err := manifest.Save(e.ManifestPath, m); err != nil {
		return results, fmt.Errorf("save manifest: %w", err)
	}
	return results, nil
}

// replaceDir copies src to dst, replacing dst if it already exists. The copy
// is recursive; we do not attempt true atomicity across the rename.
func replaceDir(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	return copyTree(src, dst)
}

func copyTree(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("copy: %s is not a directory", src)
	}
	return filepath.Walk(src, func(p string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if fi.IsDir() {
			return os.MkdirAll(target, fi.Mode()&0o777|0o700)
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink in staging tree: %s", p)
		}
		return copyFile(p, target, fi.Mode()&0o777)
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	if mode == 0 {
		mode = 0o644
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func shortSHA(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}
