// Package brain owns the Smart-Skill brain primitives used by the eval
// harness: snapshot/restore between sessions, flat-skill derivation, and
// growth stats.
//
// The brain is what makes a Smart Skill compound over time - the harness
// uses these primitives to carry state between sessions of the same arm
// so session N+1 inherits what session N wrote.
package brain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Files under references/ that are truncated to "shape only" when deriving a
// flat skill. The reason we keep headers is so tools that expect them to
// exist (lint.sh, scaffold tests) still succeed - it's the entries we strip.
var shapedMetaFiles = []string{"log.md", "patterns.md", "decisions.md", "_index.md"}

// Snapshot copies the brain portion of a skill into dst. Idempotent;
// creates dst if needed.
func Snapshot(skillDir, dst string) error {
	src := filepath.Join(skillDir, "references")
	if _, err := os.Stat(src); errors.Is(err, fs.ErrNotExist) {
		// Not a smart skill - nothing to snapshot.
		return nil
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	// Mirror references/ into dst/ - every markdown file and every
	// subtree (wiki/, raw/). Scripts and SKILL.md are NOT part of the
	// brain; the harness keeps the skill dir itself read-only so those
	// never drift.
	return copyTree(src, dst)
}

// Restore copies a previously-taken snapshot back into the live skill
// dir, replacing references/. Used to seed session N+1 from session N's
// brain-snapshot-after.
func Restore(src, skillDir string) error {
	dst := filepath.Join(skillDir, "references")
	if err := os.RemoveAll(dst); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	return copyTree(src, dst)
}

// DeriveFlat produces the flat_skill variant of src at dst. Keeps SKILL.md
// + scripts/, truncates meta files to their header sections, removes
// wiki/ and raw/. Returns the destination path so callers can pass it as
// SkillDir to a runner.
//
// When cachedSHA matches a previously-derived flat skill at the same dst,
// the copy is skipped (cheap re-runs).
func DeriveFlat(srcSkill, dst string) (string, error) {
	if err := os.RemoveAll(dst); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return "", err
	}

	// Copy SKILL.md verbatim so triggering is identical.
	if err := copyFile(filepath.Join(srcSkill, "SKILL.md"), filepath.Join(dst, "SKILL.md")); err != nil {
		return "", fmt.Errorf("copy SKILL.md: %w", err)
	}
	// Copy scripts/ so lint.sh / scaffold.sh still work when the agent
	// calls them.
	srcScripts := filepath.Join(srcSkill, "scripts")
	if _, err := os.Stat(srcScripts); err == nil {
		if err := copyTree(srcScripts, filepath.Join(dst, "scripts")); err != nil {
			return "", fmt.Errorf("copy scripts: %w", err)
		}
	}
	// Preserve references/ structure but stripped.
	refs := filepath.Join(dst, "references")
	if err := os.MkdirAll(refs, 0o755); err != nil {
		return "", err
	}
	// Keep the brain/template spec files - they're structural, not data.
	for _, name := range []string{"_brain.md", "_template.md"} {
		src := filepath.Join(srcSkill, "references", name)
		if _, err := os.Stat(src); err == nil {
			_ = copyFile(src, filepath.Join(refs, name))
		}
	}
	// Meta files: truncate to preamble (everything above the first '---'
	// separator) if there is one, else keep only the first paragraph.
	for _, name := range shapedMetaFiles {
		src := filepath.Join(srcSkill, "references", name)
		body, err := os.ReadFile(src)
		if err != nil {
			continue
		}
		out := shape(string(body))
		if err := os.WriteFile(filepath.Join(refs, name), []byte(out), 0o644); err != nil {
			return "", err
		}
	}
	// wiki/ and raw/ are omitted entirely.
	return dst, nil
}

// shape truncates a meta file to its preamble: keep everything up to the
// first `---` separator line, preserving the entry shape documentation.
// If no separator exists, keep the first paragraph.
func shape(body string) string {
	const sep = "\n---\n"
	if i := strings.Index(body, sep); i > 0 {
		return body[:i+len(sep)] + "\n(no entries yet - flat_skill arm)\n"
	}
	// First blank line as separator.
	if i := strings.Index(body, "\n\n"); i > 0 {
		return body[:i+1] + "\n(no entries yet - flat_skill arm)\n"
	}
	return body
}

// SHA returns the SHA-256 of a directory's contents (stable, sorted walk).
// Used as the cache key for DeriveFlat.
func SHA(dir string) (string, error) {
	h := sha256.New()
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		fmt.Fprintln(h, rel)
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(h, f); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Growth reports the deltas between two brain snapshots.
type Growth struct {
	WikiConcepts     Pair `json:"wiki_concepts"`
	RawFiles         Pair `json:"raw_files"`
	PatternsEntries  Pair `json:"patterns_entries"`
	DecisionsEntries Pair `json:"decisions_entries"`
	LogEntries       Pair `json:"log_entries"`
	BrainBytes       Pair `json:"brain_bytes"`
}

// Pair is the (total-now, delta-since-before) reported for every growth
// metric.
type Pair struct {
	Total int64 `json:"total"`
	Delta int64 `json:"delta"`
}

// ComputeGrowth inspects before + after snapshots and returns size / count
// deltas. Either path may be empty (means "fresh brain" for the before case).
func ComputeGrowth(beforeDir, afterDir string) (*Growth, error) {
	before, _ := inspectSnapshot(beforeDir)
	after, err := inspectSnapshot(afterDir)
	if err != nil {
		return nil, err
	}
	return &Growth{
		WikiConcepts:     pair(before.WikiConcepts, after.WikiConcepts),
		RawFiles:         pair(before.RawFiles, after.RawFiles),
		PatternsEntries:  pair(before.PatternsEntries, after.PatternsEntries),
		DecisionsEntries: pair(before.DecisionsEntries, after.DecisionsEntries),
		LogEntries:       pair(before.LogEntries, after.LogEntries),
		BrainBytes:       pair(before.BrainBytes, after.BrainBytes),
	}, nil
}

func pair(beforeVal, afterVal int64) Pair {
	return Pair{Total: afterVal, Delta: afterVal - beforeVal}
}

type snapshotStats struct {
	WikiConcepts     int64
	RawFiles         int64
	PatternsEntries  int64
	DecisionsEntries int64
	LogEntries       int64
	BrainBytes       int64
}

func inspectSnapshot(dir string) (snapshotStats, error) {
	var s snapshotStats
	if dir == "" {
		return s, nil
	}
	if _, err := os.Stat(dir); err != nil {
		return s, nil
	}
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err == nil {
			s.BrainBytes += info.Size()
		}
		rel, _ := filepath.Rel(dir, p)
		// Normalize to forward slashes so the path-prefix checks work on
		// Windows, where filepath.Rel returns `wiki\a\b\one.md`.
		rel = filepath.ToSlash(rel)
		switch {
		case strings.HasPrefix(rel, "wiki/") && strings.HasSuffix(rel, ".md"):
			s.WikiConcepts++
		case strings.HasPrefix(rel, "raw/"):
			s.RawFiles++
		}
		return nil
	})
	// Entry counts for the three journal files: count `###` headings.
	s.PatternsEntries = countHeadings(filepath.Join(dir, "patterns.md"))
	s.DecisionsEntries = countHeadings(filepath.Join(dir, "decisions.md"))
	// log.md uses `[TAG YYYY-MM-DD]` prefixes; count those.
	s.LogEntries = countLogEntries(filepath.Join(dir, "log.md"))
	return s, nil
}

func countHeadings(path string) int64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var n int64
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "### ") {
			n++
		}
	}
	return n
}

var logEntryRE = regexp.MustCompile(`^\[(INGEST|QUERY|LINT|RUN) `)

func countLogEntries(path string) int64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var n int64
	for _, line := range strings.Split(string(data), "\n") {
		if logEntryRE.MatchString(line) {
			n++
		}
	}
	return n
}

// ReadsFromBrain counts Read tool-call lines in a transcript that target
// references/*. Enables the "did the agent actually consult the brain?"
// metric without requiring runners to emit structured signals.
func ReadsFromBrain(transcript []byte) int {
	if len(transcript) == 0 {
		return 0
	}
	n := 0
	for _, line := range strings.Split(string(transcript), "\n") {
		// Match generic forms like "Read: references/patterns.md" or
		// JSONL entries containing "references/..." inside a Read tool call.
		if strings.Contains(line, "references/") &&
			(strings.Contains(line, "Read") || strings.Contains(line, "read")) {
			n++
		}
	}
	return n
}

// --- internal helpers -------------------------------------------------------

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(p, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	// Preserve mode for scripts.
	if info, err := os.Stat(src); err == nil {
		_ = os.Chmod(dst, info.Mode())
	}
	return nil
}

// MarshalJSON stabilizes the Growth shape for report snapshots.
func (g *Growth) MarshalJSON() ([]byte, error) {
	type alias Growth
	return json.Marshal((*alias)(g))
}
