// Command build-registry walks skills/, validates every SKILL.md, and writes
// registry.json at the repo root.
//
// Usage (from repo root):
//
//	go -C cli run ./cmd/build-registry
//	go -C cli run ./cmd/build-registry --check
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/v2/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/frontmatter"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/resolver"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/textutil"
)

const defaultRepo = "github.com/jjfantini/humblSKILLS"

func main() {
	var (
		skillsDir = flag.String("skills-dir", "skills", "path to the skills directory")
		outFile   = flag.String("out", "registry.json", "output registry file")
		repo      = flag.String("repo", defaultRepo, "source repo identifier")
		ref       = flag.String("ref", "", "source ref name (default: git branch or env GITHUB_REF_NAME)")
		sha       = flag.String("sha", "", "source commit sha (default: git HEAD or env GITHUB_SHA)")
		check     = flag.Bool("check", false, "exit non-zero if the generated content would differ from --out")
	)
	flag.Parse()

	if err := run(*skillsDir, *outFile, *repo, *ref, *sha, *check); err != nil {
		fmt.Fprintln(os.Stderr, "build-registry:", err)
		os.Exit(1)
	}
}

func run(skillsDir, outFile, repo, ref, sha string, check bool) error {
	adapterList, err := adapters.Load()
	if err != nil {
		return fmt.Errorf("load adapters: %w", err)
	}
	if len(adapterList) == 0 {
		return fmt.Errorf("no adapters embedded in binary")
	}

	parsed, err := walkSkills(skillsDir)
	if err != nil {
		return err
	}
	if len(parsed) == 0 {
		return fmt.Errorf("no skills found in %s", skillsDir)
	}

	known := make(map[string]string, len(parsed))
	for _, p := range parsed {
		if existing, dup := known[p.fm.Name]; dup {
			return fmt.Errorf("duplicate skill name %q (versions %s and %s)", p.fm.Name, existing, p.fm.Version())
		}
		known[p.fm.Name] = p.fm.Version()
	}

	for _, p := range parsed {
		for _, w := range p.fm.DeprecationWarnings() {
			fmt.Fprintf(os.Stderr, "build-registry: warning: %s: %s\n", p.fm.Name, w)
		}
	}

	ctx := frontmatter.ValidationContext{
		KnownSkills:   known,
		KnownAdapters: adapters.NameSet(adapterList),
	}

	var verrs []string
	for _, p := range parsed {
		if err := p.fm.Validate(p.dirName, ctx); err != nil {
			verrs = append(verrs, err.Error())
		}
	}
	if len(verrs) > 0 {
		return fmt.Errorf("skill validation failed:\n  - %s", strings.Join(verrs, "\n  - "))
	}

	if err := checkAcyclic(parsed); err != nil {
		return err
	}

	skills := make([]registry.Skill, 0, len(parsed))
	for _, p := range parsed {
		dirSha, err := registry.DirSHA(p.fullPath)
		if err != nil {
			return fmt.Errorf("dir_sha for %s: %w", p.fm.Name, err)
		}
		skills = append(skills, registry.Skill{
			Name:          p.fm.Name,
			Version:       p.fm.Version(),
			Description:   p.fm.Description,
			Category:      p.fm.Category(),
			Role:          p.fm.Role(),
			Tags:          p.fm.Tags(),
			Platforms:     p.fm.Platforms(),
			Requires:      p.fm.Requires(),
			Preserve:      p.fm.Preserve(),
			PreviousNames: p.fm.PreviousNames(),
			Upstream:      p.fm.Upstream,
			Path:          filepath.ToSlash(filepath.Join(filepath.Base(skillsDir), p.dirName)),
			DirSHA:        dirSha,
		})
	}
	sort.Slice(skills, func(i, j int) bool { return skills[i].Name < skills[j].Name })

	if ref == "" {
		ref = textutil.FirstNonEmpty(os.Getenv("GITHUB_REF_NAME"), gitBranch(), "main")
	}
	if sha == "" {
		sha = textutil.FirstNonEmpty(os.Getenv("GITHUB_SHA"), gitHeadSHA())
	}
	generatedAt := commitTime()
	if generatedAt == "" {
		generatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	reg := registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		GeneratedAt:   generatedAt,
		Source: registry.Source{
			Repo: repo,
			Ref:  ref,
			SHA:  sha,
		},
		Skills: skills,
	}

	out, err := marshalStable(reg)
	if err != nil {
		return err
	}

	if check {
		existing, err := os.ReadFile(outFile)
		if err != nil {
			return fmt.Errorf("--check could not read %s: %w", outFile, err)
		}
		diff, err := semanticDiff(existing, out)
		if err != nil {
			return fmt.Errorf("--check compare: %w", err)
		}
		if diff {
			return fmt.Errorf("%s is out of date. Run `make registry` and commit the result.", outFile)
		}
		// --check answers "would a rebuild change this file?". Content is only
		// half of that: a source.sha that predates a listed skill is also
		// something a rebuild would fix, and reporting clean there let the
		// v2.47.0 registry pass every gate while listing a skill nobody could
		// install. Fail on the same condition the write path acts on — broken
		// AND fixable — so --check never disagrees with `make registry`.
		rewrite, note := sourceSHAVerdict(existing, sha, skills)
		if note != "" {
			fmt.Fprintln(os.Stderr, "build-registry:", note)
		}
		if rewrite {
			return fmt.Errorf("%s is out of date. Run `make registry` and commit the result.", outFile)
		}
		return nil
	}

	// The generated bytes differ on every run even when no skill changed:
	// generated_at moves with the commit and source.ref/sha follow whichever
	// branch is building. Writing that unconditionally made registry.json
	// non-deterministic, and everything downstream assumed the opposite — the
	// CI auto-fix committed on every push instead of only on drift, and each
	// branch's copy conflicted with every other branch's. A PR left in a
	// conflicting state gets no workflow runs at all, so that surfaced as CI
	// silently never starting.
	//
	// Keeping the existing file when nothing semantic changed is what makes
	// "already in sync" reachable. semanticDiff is the same comparison
	// --check uses. If the existing file is unreadable or malformed, fall
	// through and write.
	//
	// The one exception is a source.sha that predates a skill the file already
	// lists — see sourceSHAVerdict. Ignoring the source block is what made the
	// early exit safe, and also what made that particular corruption permanent.
	if existing, rerr := os.ReadFile(outFile); rerr == nil {
		if diff, derr := semanticDiff(existing, out); derr == nil && !diff {
			rewrite, note := sourceSHAVerdict(existing, sha, skills)
			if note != "" {
				action := ""
				if rewrite {
					action = "; rewriting"
				}
				fmt.Fprintf(os.Stderr, "build-registry: %s%s\n", note, action)
			}
			if !rewrite {
				fmt.Printf("%s already up to date (%d skills)\n", outFile, len(skills))
				return nil
			}
		}
	}

	if err := writeAtomic(outFile, out); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%d skills)\n", outFile, len(skills))
	return nil
}

type parsedSkill struct {
	dirName  string
	fullPath string
	fm       frontmatter.Frontmatter
}

func walkSkills(skillsDir string) ([]parsedSkill, error) {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("read skills dir: %w", err)
	}
	var out []parsedSkill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirName := e.Name()
		if strings.HasPrefix(dirName, ".") {
			continue
		}
		full := filepath.Join(skillsDir, dirName)
		skillPath := filepath.Join(full, "SKILL.md")
		data, err := os.ReadFile(skillPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("read %s: %w", skillPath, err)
		}
		fm, _, err := frontmatter.Parse(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", skillPath, err)
		}
		out = append(out, parsedSkill{dirName: dirName, fullPath: full, fm: fm})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].dirName < out[j].dirName })
	return out, nil
}

func checkAcyclic(parsed []parsedSkill) error {
	g := resolver.New()
	for _, p := range parsed {
		g.AddNode(p.fm.Name)
		for _, raw := range p.fm.Requires() {
			dep, err := frontmatter.ParseDep(raw)
			if err != nil {
				return fmt.Errorf("%s: invalid dep %q: %w", p.fm.Name, raw, err)
			}
			g.AddEdge(p.fm.Name, dep.Name)
		}
	}
	if _, err := g.TopoSort(); err != nil {
		return err
	}
	return nil
}

func marshalStable(reg registry.Registry) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(reg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// semanticDiff compares two registries but ignores the generated_at and
// source fields — those track per-commit metadata (ref, sha, timestamp) that
// legitimately varies between local runs, PR builds (detached HEAD), and the
// post-merge auto-commit on main.
func semanticDiff(a, b []byte) (bool, error) {
	var ra, rb registry.Registry
	if err := json.Unmarshal(a, &ra); err != nil {
		return false, fmt.Errorf("parse existing: %w", err)
	}
	if err := json.Unmarshal(b, &rb); err != nil {
		return false, fmt.Errorf("parse new: %w", err)
	}
	ra.GeneratedAt, rb.GeneratedAt = "", ""
	ra.Source, rb.Source = registry.Source{}, registry.Source{}
	return !reflect.DeepEqual(ra, rb), nil
}

// gitPathsAt reports which of paths exist in the tree at sha, as a set. One
// `git ls-tree` call answers for every path at once; entries that do not exist
// are simply absent from the output. An error means git could not answer at
// all (not a repository, unknown or pruned object, shallow clone that never
// fetched that commit) and must never be read as "the paths are missing".
//
// Package-level so tests can substitute a fake without building a repository.
var gitPathsAt = gitLsTree

func gitLsTree(sha string, paths []string) (map[string]bool, error) {
	// --full-tree is load-bearing: without it ls-tree resolves pathspecs
	// relative to the process CWD, and the Makefile runs this binary via
	// `go -C cli run ...`, so every "skills/x" would be looked up as
	// "cli/skills/x" and report as missing. The skill paths recorded in the
	// registry are repo-root-relative, so the lookup must be too.
	args := make([]string, 0, len(paths)+6)
	args = append(args, "ls-tree", "--full-tree", "--name-only", "-z", sha, "--")
	args = append(args, paths...)

	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-tree %s: %w: %s", sha, err, strings.TrimSpace(stderr.String()))
	}

	present := make(map[string]bool, len(paths))
	for _, name := range strings.Split(string(out), "\x00") {
		if name != "" {
			present[name] = true
		}
	}
	return present, nil
}

// missingSkillAt returns the first skill path absent from the tree at sha, or
// "" when every one of them is present.
func missingSkillAt(sha string, skills []registry.Skill) (string, error) {
	paths := make([]string, 0, len(skills))
	for _, s := range skills {
		paths = append(paths, s.Path)
	}
	present, err := gitPathsAt(sha, paths)
	if err != nil {
		return "", err
	}
	for _, p := range paths {
		if !present[p] {
			return p, nil
		}
	}
	return "", nil
}

// sourceSHAVerdict decides whether a registry.json that is otherwise in sync
// must still be rewritten because its recorded source.sha predates a skill it
// lists.
//
// This is the one hole semanticDiff leaves open. `make registry` reads the
// working tree but stamps source.sha from git HEAD, so the run that first adds
// a skill necessarily records a commit that does not contain it. install
// fetches each skill's tarball at exactly that SHA, so the skill is listed and
// then fails with "no files found under skills/<name> in tarball". Because
// semanticDiff zeroes the whole source block, every later rebuild reports
// "already in sync" and the broken SHA is never replaced. It bites only the
// release that introduces a skill — for skills that already existed the frozen
// SHA still contains them, which is why it went unnoticed until v2.47.0.
//
// rewrite is true only when the recorded SHA is provably broken AND newSHA
// provably fixes it. That conjunction is what keeps this convergent: the write
// records a SHA containing every skill, so the next run takes the early exit
// and the Registry workflow pushes exactly once. Rewriting whenever the
// recorded SHA merely looked suspect would push on every trigger, forever —
// the failure mode the early exit exists to prevent.
//
// note is advisory text for the operator; it is printed whether or not the
// file is rewritten, so the unfixable case cannot fail silently.
func sourceSHAVerdict(existing []byte, newSHA string, skills []registry.Skill) (rewrite bool, note string) {
	var reg registry.Registry
	if err := json.Unmarshal(existing, &reg); err != nil {
		return false, ""
	}
	oldSHA := reg.Source.SHA
	if oldSHA == "" || newSHA == "" || oldSHA == newSHA {
		return false, ""
	}

	missing, err := missingSkillAt(oldSHA, skills)
	if err != nil {
		// Git cannot answer. Staying silent is deliberate: building outside a
		// repository (tests, a vendored tarball) is normal and must not warn.
		return false, ""
	}
	if missing == "" {
		return false, ""
	}

	if stillMissing, err := missingSkillAt(newSHA, skills); err != nil || stillMissing != "" {
		return false, fmt.Sprintf(
			"source.sha %s does not contain %s, and %s does not either — registry.json lists a skill that cannot be installed. Commit the skill, then re-run `make registry`.",
			shortSHA(oldSHA), missing, shortSHA(newSHA))
	}

	return true, fmt.Sprintf(
		"source.sha %s predates %s, which %s contains",
		shortSHA(oldSHA), missing, shortSHA(newSHA))
}

func shortSHA(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}

func writeAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func gitHeadSHA() string {
	out, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func gitBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	branch := strings.TrimSpace(string(out))
	if branch == "HEAD" {
		return ""
	}
	return branch
}

func commitTime() string {
	out, err := exec.Command("git", "log", "-1", "--format=%cI", "HEAD").Output()
	if err != nil {
		return ""
	}
	t := strings.TrimSpace(string(out))
	// %cI is strict ISO 8601; normalize to UTC for deterministic output.
	parsed, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return t
	}
	return parsed.UTC().Format(time.RFC3339)
}
