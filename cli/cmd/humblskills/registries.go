package main

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/secrets"
)

// resolvedRegistry is one registry the CLI will read, with its auth token
// already resolved.
type resolvedRegistry struct {
	Name  string
	URL   string
	Token string
}

// resolvedRegistries returns the registries to aggregate in multi-registry
// views. When the profile lists named registries, those are used (each with its
// own token); otherwise it falls back to the single resolved registry as one
// entry named "default", preserving single-registry behaviour.
func (a *App) resolvedRegistries() []resolvedRegistry {
	if p, err := profile.Load(a.Config.ProfilePath); err == nil && p != nil && len(p.Registries) > 0 {
		out := make([]resolvedRegistry, 0, len(p.Registries))
		for _, r := range p.Registries {
			tok, _ := secrets.GetRegistryTokenFor(r.Name)
			out = append(out, resolvedRegistry{Name: r.Name, URL: r.URL, Token: tok})
		}
		return out
	}
	return []resolvedRegistry{{Name: "default", URL: a.Config.RegistryURL, Token: a.registryToken()}}
}

// multiRegistry reports whether more than one registry is configured, i.e.
// whether grouped-by-registry rendering is meaningful.
func (a *App) multiRegistry() bool {
	return len(a.resolvedRegistries()) > 1
}

// fetcherForRegistry builds a Fetcher for one resolved registry, keyed to its
// own on-disk cache slot so multiple registries can coexist.
func (a *App) fetcherForRegistry(r resolvedRegistry) *registry.Fetcher {
	f := registry.NewFetcher(r.URL, a.registryCacheDir(r.Name))
	f.Token = r.Token
	return f
}

// registryCacheDir returns the cache directory for a named registry. The
// "default" fallback keeps using the base cache dir so existing caches and the
// single-registry commands stay valid.
func (a *App) registryCacheDir(name string) string {
	if name == "" || name == "default" {
		return a.Config.CacheDir
	}
	return filepath.Join(a.Config.CacheDir, "reg", sanitizeRegistryName(name))
}

func sanitizeRegistryName(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	if b.Len() == 0 {
		return "registry"
	}
	return b.String()
}

// registrySkills is the outcome of loading one registry: its full document +
// skills (tagged with the registry name) and resolved token, or the error that
// prevented loading it.
type registrySkills struct {
	Name   string
	URL    string
	Token  string
	Reg    *registry.Registry // full document, for Source (tarball fetch)
	Skills []registry.Skill   // == Reg.Skills, tagged; convenience alias
	Err    error
}

// loadRegistries loads every resolved registry, tagging each skill with its
// source registry name. Load failures are captured per registry rather than
// aborting the whole aggregation.
func (a *App) loadRegistries() []registrySkills {
	regs := a.resolvedRegistries()
	out := make([]registrySkills, 0, len(regs))
	for _, r := range regs {
		reg, _, err := a.fetcherForRegistry(r).Load()
		rs := registrySkills{Name: r.Name, URL: r.URL, Token: r.Token, Err: err}
		if err == nil && reg != nil {
			for i := range reg.Skills {
				reg.Skills[i].Registry = r.Name
			}
			rs.Reg = reg
			rs.Skills = reg.Skills
		}
		out = append(out, rs)
	}
	return out
}

// mergedRegistry flattens all successfully-loaded registries into one document
// (Source left empty — it's for listing/picking only, not fetching), sorted by
// (registry, name) so a picker renders grouped.
func mergedRegistry(loaded []registrySkills) *registry.Registry {
	skills, _ := aggregateSkills(loaded)
	return &registry.Registry{SchemaVersion: registry.SchemaVersion, Skills: skills}
}

// findRegistryForSkill returns the loaded registry that contains skillName.
func findRegistryForSkill(loaded []registrySkills, skillName string) (registrySkills, bool) {
	for _, rs := range loaded {
		for _, s := range rs.Skills {
			if s.Name == skillName {
				return rs, true
			}
		}
	}
	return registrySkills{}, false
}

// installEngineForToken builds an install Engine whose tarball fetcher carries
// the given registry's token (tarballs are SHA-keyed so the shared cache dir is
// safe across registries).
func (a *App) installEngineForToken(token string) *install.Engine {
	e := install.NewEngine(a.Config.CacheDir, a.Config.ManifestPath)
	e.Fetcher.Token = token
	return e
}

// aggregateSkills flattens loaded registries into a single slice sorted by
// (registry name, skill name), so a stable grouped order can be rendered. It
// also returns the per-registry load errors (name -> err) for reporting.
func aggregateSkills(loaded []registrySkills) ([]registry.Skill, map[string]error) {
	var all []registry.Skill
	errs := map[string]error{}
	for _, rs := range loaded {
		if rs.Err != nil {
			errs[rs.Name] = rs.Err
			continue
		}
		all = append(all, rs.Skills...)
	}
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].Registry != all[j].Registry {
			return all[i].Registry < all[j].Registry
		}
		return all[i].Name < all[j].Name
	})
	return all, errs
}
