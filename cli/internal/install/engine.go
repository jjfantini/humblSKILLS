package install

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/internal/fetch"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
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
	Skill    string `json:"skill"`
	Version  string `json:"version"`
	Platform string `json:"platform"`
	Scope    string `json:"scope"`
	Path     string `json:"path"`
	// StorePath is the canonical humblskills-owned directory Path symlinks
	// to — the source-of-truth install location, shared across every
	// platform/scope this run touched for this skill.
	StorePath string  `json:"store_path,omitempty"`
	Outcome   Outcome `json:"outcome"`
}

// Result aggregates every target outcome produced by one install run.
type Result struct {
	Results []TargetResult `json:"results"`
	// Warnings are non-fatal notices collected during the run (e.g. local
	// preserve list unreadable, fell back to registry). Every warning is also
	// emitted as a PhaseWarn event; this slice is a convenience for callers
	// that don't install an EventSink.
	Warnings []Warning `json:"warnings,omitempty"`
}

// Warning is a structured non-fatal notice from one Execute call.
type Warning struct {
	Skill    string `json:"skill,omitempty"`
	Platform string `json:"platform,omitempty"`
	Scope    string `json:"scope,omitempty"`
	Msg      string `json:"msg"`
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
	Adapters []adapters.Adapter
	// Platforms are the adapter names to install onto (order preserved).
	Platforms []string
	// Scope selects the install target scope for every adapter. Empty means
	// the adapter's DefaultScope.
	Scope string
	// Force causes existing targets to be replaced even when the manifest
	// shows they're up-to-date.
	Force bool
	// Global installs into ~/.humblskills/skills and links every selected
	// platform's user-scope target to that canonical store.
	Global bool
	// OnEvent, when set, receives progress notifications as the run proceeds.
	// Callers that don't need progress (tests, --json, scripts) can leave it
	// nil. See events.go for Phase semantics.
	OnEvent EventSink
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

	adapterIndex := make(map[string]adapters.Adapter, len(opts.Adapters))
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

	// Wrap the caller's sink so every PhaseWarn also lands in Result.Warnings.
	// Callers without an EventSink (scripts, --json) still see warnings.
	userSink := opts.OnEvent
	opts.OnEvent = func(ev Event) {
		if ev.Phase == PhaseWarn {
			res.Warnings = append(res.Warnings, Warning{
				Skill: ev.Skill, Platform: ev.Platform, Scope: ev.Scope, Msg: ev.Msg,
			})
		}
		userSink.emit(ev)
	}

	total := countTargets(plan, opts.Platforms, opts.Global)
	opts.OnEvent.emit(Event{Phase: PhaseRunStart, Total: total})

	for _, step := range plan {
		opts.OnEvent.emit(Event{
			Phase: PhaseStepStart,
			Skill: step.Skill.Name,
			IsDep: step.IsDep,
		})
		stepRes, err := e.installOne(reg, step, opts, adapterIndex, m)
		if err != nil {
			opts.OnEvent.emit(Event{
				Phase: PhaseError,
				Skill: step.Skill.Name,
				Err:   err,
			})
			return res, fmt.Errorf("%s: %w", step.Skill.Name, err)
		}
		res.Results = append(res.Results, stepRes...)
	}

	if err := manifest.Save(e.ManifestPath, m); err != nil {
		opts.OnEvent.emit(Event{Phase: PhaseError, Err: err})
		return res, fmt.Errorf("save manifest: %w", err)
	}
	opts.OnEvent.emit(Event{Phase: PhaseRunDone, Total: total})
	return res, nil
}

// countTargets walks plan and totals the (skill, platform, scope) triples the
// run will visit, honouring each skill's platforms[] allow-list. Cheap, keeps
// progress denominators honest without guesswork.
func countTargets(plan []Step, platforms []string, global bool) int {
	total := 0
	for _, step := range plan {
		allow := map[string]struct{}{}
		if global || len(step.Skill.Platforms) == 0 {
			for _, p := range platforms {
				allow[p] = struct{}{}
			}
		} else {
			for _, p := range step.Skill.Platforms {
				allow[p] = struct{}{}
			}
		}
		for _, p := range platforms {
			if _, ok := allow[p]; ok {
				total++
			}
		}
	}
	return total
}

type installPending struct {
	adapter adapters.Adapter
	target  adapters.Target
	final   string
}

type installPlanTarget struct {
	pending      installPending
	out          Outcome
	orphan       string
	existingPath string
}

func (e *Engine) installOne(
	reg *registry.Registry,
	step Step,
	opts ExecuteOpts,
	adapterIndex map[string]adapters.Adapter,
	m *manifest.Manifest,
) ([]TargetResult, error) {
	skill := step.Skill

	allow := map[string]struct{}{}
	if opts.Global || len(skill.Platforms) == 0 {
		for _, p := range opts.Platforms {
			allow[p] = struct{}{}
		}
	} else {
		for _, p := range skill.Platforms {
			allow[p] = struct{}{}
		}
	}

	var pendings []installPending
	for _, p := range opts.Platforms {
		if _, ok := allow[p]; !ok {
			continue
		}
		adapter := adapterIndex[p]
		scope := opts.Scope
		if opts.Global {
			scope = "user"
		} else if scope == "" {
			scope = adapter.DefaultScope
		}
		t, err := adapter.Target(scope)
		if err != nil {
			return nil, err
		}
		pendings = append(pendings, installPending{
			adapter: adapter,
			target:  t,
			final:   filepath.Join(t.Path, skill.Name),
		})
	}
	if len(pendings) == 0 {
		return nil, nil
	}

	storePath, err := CanonicalSkillPath(skill.Name, pendings[0].target.Scope, opts.Global)
	if err != nil {
		return nil, err
	}
	mode := installMode(opts.Global)

	var toWrite []installPlanTarget
	var skipped []TargetResult
	preserveSource := ""
	for _, pg := range pendings {
		existing := m.FindOne(skill.Name, pg.adapter.Name, pg.target.Scope)
		orphan := ""
		if existing != nil && existing.Path != pg.final {
			orphan = existing.Path
		}

		upToDate := existing != nil &&
			existing.Version == skill.Version &&
			existing.RegistryRef == skill.DirSHA &&
			existing.Path == pg.final &&
			existing.StorePath == storePath &&
			existing.InstallMode == mode
		if !targetLinksToStore(pg.final, storePath) {
			upToDate = false
		}
		if _, err := os.Stat(filepath.Join(storePath, "SKILL.md")); err != nil {
			upToDate = false
		}

		existingPath := ""
		if existing != nil {
			existingPath = existing.Path
			if existing.StorePath != "" {
				preserveSource = existing.StorePath
			} else if preserveSource == "" {
				preserveSource = existing.Path
			}
		} else if preserveSource == "" && realDirExists(pg.final) {
			preserveSource = pg.final
		}
		switch {
		case upToDate && !opts.Force:
			opts.OnEvent.emit(Event{
				Phase: PhaseTargetStart, Skill: skill.Name, IsDep: step.IsDep,
				Platform: pg.adapter.Name, Scope: pg.target.Scope,
			})
			skipped = append(skipped, TargetResult{
				Skill: skill.Name, Version: skill.Version, Platform: pg.adapter.Name, Scope: pg.target.Scope,
				Path: pg.final, StorePath: storePath, Outcome: OutcomeSkipped,
			})
			opts.OnEvent.emit(Event{
				Phase: PhaseTargetDone, Skill: skill.Name, IsDep: step.IsDep,
				Platform: pg.adapter.Name, Scope: pg.target.Scope, Outcome: OutcomeSkipped,
				Path: pg.final, Version: skill.Version, StorePath: storePath,
			})
		case upToDate && opts.Force:
			toWrite = append(toWrite, installPlanTarget{pending: pg, out: OutcomeForced, orphan: orphan, existingPath: existingPath})
		case existing != nil:
			toWrite = append(toWrite, installPlanTarget{pending: pg, out: OutcomeReplaced, orphan: orphan, existingPath: existingPath})
		default:
			toWrite = append(toWrite, installPlanTarget{pending: pg, out: OutcomeInstalled, orphan: orphan, existingPath: existingPath})
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

	writeSrc := staging
	if preserveSource != "" && !opts.Force {
		preserveList, userOwnsPreserve, err := e.preserveListForStore(skill, preserveSource, toWrite, opts)
		if err != nil {
			return nil, err
		}
		if len(preserveList) > 0 || userOwnsPreserve {
			perStore := filepath.Join(e.StagingDir, skill.Name+"-"+shortSHA(reg.Source.SHA)+"-store")
			if err := os.RemoveAll(perStore); err != nil {
				return nil, fmt.Errorf("clean store staging: %w", err)
			}
			defer os.RemoveAll(perStore)
			if err := copyTree(staging, perStore); err != nil {
				return nil, fmt.Errorf("copy store staging: %w", err)
			}
			if len(preserveList) > 0 {
				if err := applyPreserve(preserveSource, perStore, preserveList); err != nil {
					return nil, err
				}
			}
			if userOwnsPreserve {
				skillMDPath := filepath.Join(perStore, "SKILL.md")
				if err := mergePreserveIntoSkillMD(skillMDPath, preserveList); err != nil {
					return nil, fmt.Errorf("merge preserve into SKILL.md: %w", err)
				}
			}
			writeSrc = perStore
		}
	}

	if err := replaceDir(writeSrc, storePath); err != nil {
		return nil, fmt.Errorf("place canonical store %s: %w", storePath, err)
	}

	results := append([]TargetResult(nil), skipped...)
	for _, pt := range toWrite {
		opts.OnEvent.emit(Event{
			Phase: PhaseTargetStart, Skill: skill.Name, IsDep: step.IsDep,
			Platform: pt.pending.adapter.Name, Scope: pt.pending.target.Scope,
		})
		if pt.orphan != "" && pt.orphan != pt.pending.final {
			if err := os.RemoveAll(pt.orphan); err != nil {
				return nil, fmt.Errorf("clean old install %s: %w", pt.orphan, err)
			}
		}
		if err := linkStore(storePath, pt.pending.final); err != nil {
			return nil, fmt.Errorf("link %s -> %s: %w", pt.pending.final, storePath, err)
		}
		m.Upsert(manifest.Installation{
			Skill:       skill.Name,
			Version:     skill.Version,
			Platform:    pt.pending.adapter.Name,
			Scope:       pt.pending.target.Scope,
			Path:        pt.pending.final,
			StorePath:   storePath,
			InstallMode: mode,
			InstalledAt: e.Now().UTC(),
			SourceSHA:   reg.Source.SHA,
			RegistryRef: skill.DirSHA,
		})
		results = append(results, TargetResult{
			Skill:     skill.Name,
			Version:   skill.Version,
			Platform:  pt.pending.adapter.Name,
			Scope:     pt.pending.target.Scope,
			Path:      pt.pending.final,
			StorePath: storePath,
			Outcome:   pt.out,
		})
		opts.OnEvent.emit(Event{
			Phase: PhaseTargetDone, Skill: skill.Name, IsDep: step.IsDep,
			Platform: pt.pending.adapter.Name, Scope: pt.pending.target.Scope,
			Outcome: pt.out, Path: pt.pending.final, Version: skill.Version, StorePath: storePath,
		})
	}
	return results, nil
}

func (e *Engine) preserveListForStore(
	skill registry.Skill,
	preserveSource string,
	targets []installPlanTarget,
	opts ExecuteOpts,
) ([]string, bool, error) {
	if local, ok, reason := loadLocalPreserve(preserveSource); ok {
		return local, true, nil
	} else {
		for _, pt := range targets {
			opts.OnEvent.emit(Event{
				Phase: PhaseWarn, Skill: skill.Name,
				Platform: pt.pending.adapter.Name, Scope: pt.pending.target.Scope,
				Msg: fmt.Sprintf("local preserve unreadable (%s); falling back to registry list", reason),
			})
		}
	}
	return skill.Preserve, false, nil
}

func targetLinksToStore(targetPath, storePath string) bool {
	link, err := os.Readlink(targetPath)
	if err != nil {
		return false
	}
	if !filepath.IsAbs(link) {
		link = filepath.Join(filepath.Dir(targetPath), link)
	}
	return filepath.Clean(link) == filepath.Clean(storePath)
}

func realDirExists(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return false
	}
	return info.IsDir()
}

func linkStore(storePath, targetPath string) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	if err := os.RemoveAll(targetPath); err != nil {
		return err
	}
	return os.Symlink(storePath, targetPath)
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
	storeRefs := map[string]struct{}{}
	for _, inst := range entries {
		if inst.StorePath != "" {
			storeRefs[inst.StorePath] = struct{}{}
		}
		if err := os.RemoveAll(inst.Path); err != nil {
			return results, fmt.Errorf("remove %s: %w", inst.Path, err)
		}
		results = append(results, TargetResult{
			Skill: inst.Skill, Platform: inst.Platform, Scope: inst.Scope,
			Path: inst.Path, Outcome: "removed",
		})
	}
	m.Remove(skill)
	for storePath := range storeRefs {
		if !manifestReferencesStore(m, storePath) {
			if err := os.RemoveAll(storePath); err != nil {
				return results, fmt.Errorf("remove canonical store %s: %w", storePath, err)
			}
		}
	}
	if err := manifest.Save(e.ManifestPath, m); err != nil {
		return results, fmt.Errorf("save manifest: %w", err)
	}
	return results, nil
}

func manifestReferencesStore(m *manifest.Manifest, storePath string) bool {
	for _, inst := range m.Installations {
		if inst.StorePath == storePath {
			return true
		}
	}
	return false
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

// applyPreserve merges user-owned content from userRoot into stagingRoot
// according to the preserve list.
//
//   - File entry (no trailing "/"): user wins; user's bytes overwrite whatever
//     staging shipped.
//   - Directory entry (trailing "/"): deep merge; staging wins on per-file
//     conflicts. Files only in user land alongside staging's version.
//
// Type mismatches (file entry vs user dir, or dir entry vs user file) and
// symlinks in the user source are rejected rather than silently resolved.
func applyPreserve(userRoot, stagingRoot string, entries []string) error {
	for _, raw := range entries {
		rel := strings.TrimSpace(raw)
		rel = strings.TrimPrefix(rel, "./")
		isDir := strings.HasSuffix(rel, "/")
		relClean := filepath.FromSlash(strings.TrimSuffix(rel, "/"))
		if relClean == "" || relClean == "." {
			continue
		}
		srcAbs := filepath.Join(userRoot, relClean)
		dstAbs := filepath.Join(stagingRoot, relClean)

		fi, err := os.Lstat(srcAbs)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("preserve stat %s: %w", srcAbs, err)
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("preserve: refusing to follow symlink %s", srcAbs)
		}

		if isDir {
			if !fi.IsDir() {
				return fmt.Errorf("preserve: entry %q declares a directory but %s is a file", raw, srcAbs)
			}
			if err := preserveMergeDir(srcAbs, dstAbs); err != nil {
				return err
			}
			continue
		}
		if fi.IsDir() {
			return fmt.Errorf("preserve: entry %q declares a file but %s is a directory", raw, srcAbs)
		}
		if err := copyFile(srcAbs, dstAbs, fi.Mode()&0o777); err != nil {
			return fmt.Errorf("preserve copy %s: %w", srcAbs, err)
		}
	}
	return nil
}

// preserveMergeDir walks userDir and copies every regular file into dstDir
// only when dstDir does not already have that relative path — staging wins on
// conflicts. Symlinks anywhere in userDir abort the merge.
func preserveMergeDir(userDir, dstDir string) error {
	return filepath.Walk(userDir, func(p string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(userDir, p)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dstDir, fi.Mode()&0o777|0o700)
		}
		dst := filepath.Join(dstDir, rel)
		if fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("preserve: refusing to follow symlink %s", p)
		}
		if fi.IsDir() {
			if _, err := os.Stat(dst); err == nil {
				return nil
			}
			return os.MkdirAll(dst, fi.Mode()&0o777|0o700)
		}
		if _, err := os.Stat(dst); err == nil {
			// Staging already ships this file; staging wins.
			return nil
		}
		return copyFile(p, dst, fi.Mode()&0o777)
	})
}
