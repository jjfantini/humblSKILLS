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

	"github.com/jjfantini/humblSKILLS/cli/internal/frontmatter"
	"github.com/jjfantini/humblSKILLS/cli/internal/platform"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/resolver"
)

const defaultRepo = "github.com/jjfantini/humblSKILLS"

func main() {
	var (
		skillsDir   = flag.String("skills-dir", "skills", "path to the skills directory")
		adaptersDir = flag.String("adapters-dir", "adapters", "path to the adapters directory")
		outFile     = flag.String("out", "registry.json", "output registry file")
		repo        = flag.String("repo", defaultRepo, "source repo identifier")
		ref         = flag.String("ref", "", "source ref name (default: git branch or env GITHUB_REF_NAME)")
		sha         = flag.String("sha", "", "source commit sha (default: git HEAD or env GITHUB_SHA)")
		check       = flag.Bool("check", false, "exit non-zero if the generated content would differ from --out")
	)
	flag.Parse()

	if err := run(*skillsDir, *adaptersDir, *outFile, *repo, *ref, *sha, *check); err != nil {
		fmt.Fprintln(os.Stderr, "build-registry:", err)
		os.Exit(1)
	}
}

func run(skillsDir, adaptersDir, outFile, repo, ref, sha string, check bool) error {
	adapters, err := platform.LoadAll(adaptersDir)
	if err != nil {
		return fmt.Errorf("load adapters: %w", err)
	}
	if len(adapters) == 0 {
		return fmt.Errorf("no adapters found in %s", adaptersDir)
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
			return fmt.Errorf("duplicate skill name %q (versions %s and %s)", p.fm.Name, existing, p.fm.Version)
		}
		known[p.fm.Name] = p.fm.Version
	}

	ctx := frontmatter.ValidationContext{
		KnownSkills:   known,
		KnownAdapters: platform.NameSet(adapters),
	}

	var verrs []string
	for _, p := range parsed {
		if err := p.fm.Validate(p.dirName, p.fullPath, ctx); err != nil {
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
			Name:        p.fm.Name,
			Version:     p.fm.Version,
			Description: p.fm.Description,
			Tags:        p.fm.Tags,
			Platforms:   p.fm.Platforms,
			Requires:    p.fm.Requires,
			Path:        filepath.ToSlash(filepath.Join(filepath.Base(skillsDir), p.dirName)),
			PostInstall: p.fm.PostInstall,
			DirSHA:      dirSha,
		})
	}
	sort.Slice(skills, func(i, j int) bool { return skills[i].Name < skills[j].Name })

	if ref == "" {
		ref = firstNonEmpty(os.Getenv("GITHUB_REF_NAME"), gitBranch(), "main")
	}
	if sha == "" {
		sha = firstNonEmpty(os.Getenv("GITHUB_SHA"), gitHeadSHA())
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
		return nil
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
		for _, raw := range p.fm.Requires {
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

// semanticDiff compares two registries but ignores generated_at and
// source.sha — those fields track metadata that changes each commit.
func semanticDiff(a, b []byte) (bool, error) {
	var ra, rb registry.Registry
	if err := json.Unmarshal(a, &ra); err != nil {
		return false, fmt.Errorf("parse existing: %w", err)
	}
	if err := json.Unmarshal(b, &rb); err != nil {
		return false, fmt.Errorf("parse new: %w", err)
	}
	ra.GeneratedAt, rb.GeneratedAt = "", ""
	ra.Source.SHA, rb.Source.SHA = "", ""
	return !reflect.DeepEqual(ra, rb), nil
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

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
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
