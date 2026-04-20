// Package workspace owns the on-disk layout for eval runs.
//
// Layout (under <root>/<skill>/):
//
//	iterations.json       registry of iterations with headline stats
//	iteration-<N>/        one run; immutable once marked complete
//	  <config>/           arm (smart_skill, flat_skill, no_skill)
//	    session-<M>/      one session of a scenario
//	      eval-<id>/
//	        outputs/
//	        transcript.txt
//	        timing.json
//	        metrics.json
//	        grading.json
//	    brain-snapshot-before/
//	    brain-snapshot-after/
//	  benchmark.json
//	  trajectory.json
//	  growth.json
//	  report.{html,md,json}
//
// Iterations are persistent; nothing here is cache-grade. Use Prune to drop
// old ones.
package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/adrg/xdg"
)

// SchemaVersion is the current iterations.json schema.
const SchemaVersion = 1

// Status values for an iteration.
const (
	StatusRunning  = "running"
	StatusComplete = "complete"
	StatusFailed   = "failed"
	StatusAborted  = "aborted"
)

// Registry is the iterations.json document.
type Registry struct {
	SchemaVersion int         `json:"schema_version"`
	Iterations    []Iteration `json:"iterations"`
}

// Iteration is one entry in the registry.
type Iteration struct {
	N              int               `json:"n"`
	StartedAt      time.Time         `json:"started_at"`
	CompletedAt    *time.Time        `json:"completed_at,omitempty"`
	Status         string            `json:"status"`
	Runner         string            `json:"runner"`
	Arms           []string          `json:"arms"`
	Scenarios      []string          `json:"scenarios"`
	HeadlinePassRt map[string]float64 `json:"headline_pass_rate,omitempty"`
	Tokens         map[string]int    `json:"tokens,omitempty"`
	RegradedFrom   int               `json:"regraded_from,omitempty"`
}

// Resolver centralizes the root-path precedence (flag > env > profile > XDG).
type Resolver struct {
	// FlagOverride is --workspace from the CLI, if set.
	FlagOverride string
	// EnvOverride is HUMBLSKILLS_EVAL_WORKSPACE, if set.
	EnvOverride string
	// ProfileDefault is the profile's eval.default_workspace, if set.
	ProfileDefault string
}

// Root returns the absolute root path. Never returns empty on success.
func (r Resolver) Root() (string, error) {
	for _, cand := range []string{r.FlagOverride, r.EnvOverride, r.ProfileDefault} {
		if cand != "" {
			abs, err := filepath.Abs(cand)
			if err != nil {
				return "", err
			}
			return abs, nil
		}
	}
	return DefaultRoot()
}

// DefaultRoot is the XDG_STATE_HOME/humblskills/evals path. Exposed so doctor
// and CLI can report it.
func DefaultRoot() (string, error) {
	if xdg.StateHome != "" {
		return filepath.Join(xdg.StateHome, "humblskills", "evals"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve workspace root: %w", err)
	}
	return filepath.Join(home, ".local", "state", "humblskills", "evals"), nil
}

// SkillDir returns <root>/<skill>/.
func SkillDir(root, skill string) string { return filepath.Join(root, skill) }

// RegistryPath returns <root>/<skill>/iterations.json.
func RegistryPath(root, skill string) string {
	return filepath.Join(SkillDir(root, skill), "iterations.json")
}

// IterationDir returns <root>/<skill>/iteration-<N>/.
func IterationDir(root, skill string, n int) string {
	return filepath.Join(SkillDir(root, skill), fmt.Sprintf("iteration-%d", n))
}

// LoadRegistry reads the registry or returns an empty one if missing.
func LoadRegistry(root, skill string) (*Registry, error) {
	path := RegistryPath(root, skill)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Registry{SchemaVersion: SchemaVersion}, nil
		}
		return nil, fmt.Errorf("read iterations.json: %w", err)
	}
	var r Registry
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse iterations.json: %w", err)
	}
	if r.SchemaVersion == 0 {
		r.SchemaVersion = SchemaVersion
	}
	if r.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("unsupported iterations schema_version %d", r.SchemaVersion)
	}
	return &r, nil
}

// SaveRegistry writes the registry atomically.
func SaveRegistry(root, skill string, r *Registry) error {
	if r == nil {
		return errors.New("nil registry")
	}
	r.SchemaVersion = SchemaVersion
	path := RegistryPath(root, skill)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// NextIterationN walks iteration-* directories and returns max+1.
// Used once per eval run to allocate a fresh, monotonic iteration number.
func NextIterationN(root, skill string) (int, error) {
	max, err := MaxIterationN(root, skill)
	if err != nil {
		return 0, err
	}
	return max + 1, nil
}

// MaxIterationN returns the highest iteration number on disk, or 0 if none.
func MaxIterationN(root, skill string) (int, error) {
	skillDir := SkillDir(root, skill)
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	max := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if !strings.HasPrefix(e.Name(), "iteration-") {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(e.Name(), "iteration-"))
		if err != nil {
			continue
		}
		if n > max {
			max = n
		}
	}
	return max, nil
}

// BeginIteration allocates a new iteration, creates its directory, and
// appends a running entry to the registry. Returns the iteration number +
// its absolute path.
func BeginIteration(root, skill string, runner string, arms, scenarios []string) (int, string, error) {
	n, err := NextIterationN(root, skill)
	if err != nil {
		return 0, "", err
	}
	dir := IterationDir(root, skill, n)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, "", err
	}
	reg, err := LoadRegistry(root, skill)
	if err != nil {
		return 0, "", err
	}
	reg.Iterations = append(reg.Iterations, Iteration{
		N:         n,
		StartedAt: time.Now().UTC(),
		Status:    StatusRunning,
		Runner:    runner,
		Arms:      arms,
		Scenarios: scenarios,
	})
	if err := SaveRegistry(root, skill, reg); err != nil {
		return 0, "", err
	}
	return n, dir, nil
}

// CompleteIteration stamps the iteration complete and records headline stats.
func CompleteIteration(root, skill string, n int, passRates map[string]float64, tokens map[string]int) error {
	reg, err := LoadRegistry(root, skill)
	if err != nil {
		return err
	}
	for i := range reg.Iterations {
		if reg.Iterations[i].N == n {
			t := time.Now().UTC()
			reg.Iterations[i].CompletedAt = &t
			reg.Iterations[i].Status = StatusComplete
			reg.Iterations[i].HeadlinePassRt = passRates
			reg.Iterations[i].Tokens = tokens
			return SaveRegistry(root, skill, reg)
		}
	}
	return fmt.Errorf("iteration %d not found in registry", n)
}

// MarkIteration updates the status of an iteration (failed, aborted).
func MarkIteration(root, skill string, n int, status string) error {
	reg, err := LoadRegistry(root, skill)
	if err != nil {
		return err
	}
	for i := range reg.Iterations {
		if reg.Iterations[i].N == n {
			reg.Iterations[i].Status = status
			return SaveRegistry(root, skill, reg)
		}
	}
	return fmt.Errorf("iteration %d not found in registry", n)
}

// ListSkills enumerates skill subdirs under root.
func ListSkills(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}

// SizeBytes sums file sizes under path. Used by the `ls` / `prune` commands
// and the doctor workspace check.
func SizeBytes(path string) (int64, error) {
	var total int64
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		total += info.Size()
		return nil
	})
	if errors.Is(err, fs.ErrNotExist) {
		return 0, nil
	}
	return total, err
}

// PruneOpts controls Prune behaviour.
type PruneOpts struct {
	KeepLast   int           // retain the N most recent iterations (0 = keep everything)
	OlderThan  time.Duration // drop iterations whose StartedAt is older than this (0 = no time filter)
	DryRun     bool
	All        bool // drop every iteration (requires explicit opt-in from caller)
}

// PruneResult reports what Prune did (or would do in DryRun).
type PruneResult struct {
	Removed    []int // iteration numbers removed
	BytesFreed int64
}

// Prune applies the retention policy to a skill's workspace and updates the
// registry. Immutable iterations still get removed - pruning is deletion.
func Prune(root, skill string, opts PruneOpts) (*PruneResult, error) {
	reg, err := LoadRegistry(root, skill)
	if err != nil {
		return nil, err
	}
	// Sort by N descending so KeepLast is easy to apply.
	sort.Slice(reg.Iterations, func(i, j int) bool {
		return reg.Iterations[i].N > reg.Iterations[j].N
	})
	now := time.Now()
	res := &PruneResult{}
	var keep []Iteration
	for i, it := range reg.Iterations {
		drop := false
		switch {
		case opts.All:
			drop = true
		case opts.KeepLast > 0 && i >= opts.KeepLast:
			drop = true
		case opts.OlderThan > 0 && now.Sub(it.StartedAt) > opts.OlderThan:
			drop = true
		}
		if !drop {
			keep = append(keep, it)
			continue
		}
		dir := IterationDir(root, skill, it.N)
		size, _ := SizeBytes(dir)
		res.BytesFreed += size
		res.Removed = append(res.Removed, it.N)
		if !opts.DryRun {
			if err := os.RemoveAll(dir); err != nil {
				return res, fmt.Errorf("remove iteration-%d: %w", it.N, err)
			}
		}
	}
	// Restore natural ascending order for the saved registry.
	sort.Slice(keep, func(i, j int) bool { return keep[i].N < keep[j].N })
	reg.Iterations = keep
	if !opts.DryRun {
		if err := SaveRegistry(root, skill, reg); err != nil {
			return res, err
		}
	}
	sort.Ints(res.Removed)
	return res, nil
}

// HumanSize formats a byte count with a binary SI-style suffix. Used by the
// TUI `ls` and `prune --dry-run` output.
func HumanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
