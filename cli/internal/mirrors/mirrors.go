// Package mirrors detects drift between a mirrored skill and its upstream.
//
// The split this package enforces: detection is deterministic and automatable,
// re-distillation is not. Nothing here rewrites a skill. The most it produces
// is a work order (see [Plan]) for a human or agent to execute.
//
// Two invariants make that safe:
//
//   - The preserved copy under references/raw/ IS the drift baseline. Current
//     upstream is compared against it byte-for-byte, so no hash or lockfile is
//     needed - but it also means nothing may silently overwrite that file, or
//     the next comparison has nothing to compare against.
//   - The distillation in references/wiki/ is the product. It is never
//     auto-merged.
package mirrors

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/v2/internal/frontmatter"
)

// Status is the drift severity for one mirrored skill.
//
// The distinction between Drifted and Rewritten is the point: a typo fix and a
// wholesale rewrite must not produce the same alert, or the alert gets ignored.
type Status string

const (
	// StatusCurrent means the preserved copy matches upstream byte-for-byte.
	StatusCurrent Status = "current"
	// StatusDrifted means upstream changed but its structure is recognisable -
	// a spot check against the diff is usually enough.
	StatusDrifted Status = "drifted"
	// StatusRewritten means upstream's heading structure has substantially
	// changed. Concepts likely need re-distilling, not patching.
	StatusRewritten Status = "rewritten"
	// StatusUnknown means drift could not be determined (no fetch source, or
	// the fetch failed). Never treated as "fine".
	StatusUnknown Status = "unknown"
)

// rewriteThreshold is the share of upstream headings that must still be present
// in the preserved copy for a change to count as Drifted rather than Rewritten.
const rewriteThreshold = 0.5

// Result is the outcome of checking one mirrored skill.
type Result struct {
	Skill     string
	Dir       string
	Upstream  *frontmatter.Upstream
	Status    Status
	Reason    string
	Preserved []byte
	Current   []byte
	// SharedHeadings / UpstreamHeadings quantify structural overlap and are
	// what separate Drifted from Rewritten.
	SharedHeadings   int
	UpstreamHeadings int
}

// Stale reports whether this result needs human attention.
func (r Result) Stale() bool {
	return r.Status == StatusDrifted || r.Status == StatusRewritten
}

// BaselineFile resolves an Upstream.Preserved value to the single file used as
// the drift baseline.
//
// A skill that mirrors several upstream files (the better-* family preserves a
// SKILL.md plus companions) declares the directory with a trailing slash. The
// upstream SKILL.md inside it is the canonical artifact: it is what defines the
// skill, so it is what drift is measured against.
func BaselineFile(preserved string) string {
	if strings.HasSuffix(preserved, "/") {
		return preserved + "SKILL.md"
	}
	return preserved
}

var headingRe = regexp.MustCompile(`(?m)^#{1,6}[ \t]+(.+?)[ \t]*$`)

func headings(b []byte) []string {
	var out []string
	for _, m := range headingRe.FindAllSubmatch(b, -1) {
		out = append(out, strings.ToLower(strings.TrimSpace(string(m[1]))))
	}
	return out
}

// Fetcher retrieves the current upstream bytes for an Upstream declaration.
type Fetcher func(ctx context.Context, u frontmatter.Upstream) ([]byte, error)

// DefaultFetcher reads from an https:// URL or a local path. A local path is
// how the Anthropic plugin skills are reached - they are already on disk in the
// plugin cache, so checking them costs no network at all.
func DefaultFetcher(ctx context.Context, u frontmatter.Upstream) ([]byte, error) {
	if u.Fetch == "" {
		return nil, fmt.Errorf("no fetch source declared")
	}
	if strings.HasPrefix(u.Fetch, "http://") || strings.HasPrefix(u.Fetch, "https://") {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.Fetch, nil)
		if err != nil {
			return nil, err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("upstream returned %s", resp.Status)
		}
		return io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	}
	path := u.Fetch
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, path[2:])
	}
	return os.ReadFile(path)
}

// Check inspects every skill under skillsDir that declares an `upstream:`
// block. Skills without one are not mirrors and are skipped silently.
//
// A fetch failure yields StatusUnknown rather than an error: one unreachable
// upstream must not hide drift in the others.
func Check(ctx context.Context, skillsDir string, fetch Fetcher) ([]Result, error) {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, err
	}
	if fetch == nil {
		fetch = DefaultFetcher
	}

	var out []Result
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(skillsDir, e.Name())
		raw, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
		if err != nil {
			continue
		}
		fm, _, err := frontmatter.Parse(raw)
		if err != nil || fm.Upstream == nil {
			continue
		}
		out = append(out, checkOne(ctx, dir, fm.Name, *fm.Upstream, fetch))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Skill < out[j].Skill })
	return out, nil
}

func checkOne(ctx context.Context, dir, name string, u frontmatter.Upstream, fetch Fetcher) Result {
	r := Result{Skill: name, Dir: dir, Upstream: &u}

	if u.Preserved == "" {
		r.Status, r.Reason = StatusUnknown, "upstream block declares no preserved: copy"
		return r
	}
	preserved, err := os.ReadFile(filepath.Join(dir, BaselineFile(u.Preserved)))
	if err != nil {
		r.Status, r.Reason = StatusUnknown, fmt.Sprintf("preserved copy unreadable: %v", err)
		return r
	}
	r.Preserved = preserved

	current, err := fetch(ctx, u)
	if err != nil {
		r.Status, r.Reason = StatusUnknown, fmt.Sprintf("could not read upstream: %v", err)
		return r
	}
	r.Current = current

	if string(normalise(preserved)) == string(normalise(current)) {
		r.Status, r.Reason = StatusCurrent, "preserved copy matches upstream"
		return r
	}

	pre, cur := headings(preserved), headings(current)
	have := make(map[string]bool, len(pre))
	for _, h := range pre {
		have[h] = true
	}
	for _, h := range cur {
		if have[h] {
			r.SharedHeadings++
		}
	}
	r.UpstreamHeadings = len(cur)

	share := 1.0
	if r.UpstreamHeadings > 0 {
		share = float64(r.SharedHeadings) / float64(r.UpstreamHeadings)
	}
	if share < rewriteThreshold {
		r.Status = StatusRewritten
		r.Reason = fmt.Sprintf("upstream restructured - only %d of %d headings survive; concepts likely need re-distilling",
			r.SharedHeadings, r.UpstreamHeadings)
		return r
	}
	r.Status = StatusDrifted
	r.Reason = fmt.Sprintf("upstream edited - %d of %d headings unchanged; review the diff",
		r.SharedHeadings, r.UpstreamHeadings)
	return r
}

// normalise strips trailing whitespace and normalises line endings so a
// CRLF or trailing-newline difference is not reported as drift.
func normalise(b []byte) []byte {
	s := strings.ReplaceAll(string(b), "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return []byte(strings.TrimRight(strings.Join(lines, "\n"), "\n"))
}

// Summary renders a one-line status suitable for a CLI banner. Empty when
// nothing needs attention, so callers can print it unconditionally.
func Summary(results []Result) string {
	var stale, unknown int
	for _, r := range results {
		switch {
		case r.Stale():
			stale++
		case r.Status == StatusUnknown:
			unknown++
		}
	}
	switch {
	case stale > 0 && unknown > 0:
		return fmt.Sprintf("%d mirrored skill(s) drifted from upstream, %d uncheckable - run `humblskills mirrors check`", stale, unknown)
	case stale > 0:
		return fmt.Sprintf("%d mirrored skill(s) drifted from upstream - run `humblskills mirrors check`", stale)
	case unknown > 0:
		return fmt.Sprintf("%d mirrored skill(s) could not be checked - run `humblskills mirrors check`", unknown)
	}
	return ""
}

// CacheTTL is how long a check result stays fresh for opportunistic callers.
// Upstream skills change on the order of weeks; a daily check is already
// generous, and anything shorter taxes every command for no signal.
const CacheTTL = 24 * time.Hour

// ShouldCheck reports whether a background/opportunistic check is due, based on
// the mtime of a stamp file. Callers that hit the network anyway (install,
// update, search) can gate on this so offline commands stay instant.
func ShouldCheck(stampPath string) bool {
	info, err := os.Stat(stampPath)
	if err != nil {
		return true
	}
	return time.Since(info.ModTime()) > CacheTTL
}

// Stamp records that a check just ran.
func Stamp(stampPath string) error {
	if err := os.MkdirAll(filepath.Dir(stampPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(stampPath, []byte(time.Now().UTC().Format(time.RFC3339)+"\n"), 0o644)
}
