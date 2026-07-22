package main

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/textutil"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
)

func newListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed skills",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runList(app, false)
		},
	}
}

// runList is the package-level entry point. `fromDashboard` softens ESC/quit
// semantics so the launcher loop in start.go can re-enter the dashboard
// instead of exiting the process.
func runList(app *App, fromDashboard bool) error {
	m, err := manifest.Load(app.Config.ManifestPath)
	if err != nil {
		return err
	}

	// Best-effort, offline-only drift lookup: `list` reports the last-known
	// registry state without forcing a network fetch, so it stays fast and
	// works offline. `update` / `registry refresh` are what refill the cache.
	avail := availableUpdates(app, m)

	if app.Config.JSON {
		return app.UI.JSON(buildListView(m, avail))
	}
	if len(m.Installations) == 0 {
		app.UI.Info("no skills installed — try 'humblskills install <name>'")
		return nil
	}
	installs := append([]manifest.Installation(nil), m.Installations...)
	sort.Slice(installs, func(i, j int) bool {
		if installs[i].Skill != installs[j].Skill {
			return installs[i].Skill < installs[j].Skill
		}
		if installs[i].Platform != installs[j].Platform {
			return installs[i].Platform < installs[j].Platform
		}
		return installs[i].Scope < installs[j].Scope
	})

	if tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes) {
		return runListTUI(app, m, fromDashboard)
	}
	renderListTable(app, installs, avail)
	return nil
}

// availableUpdates maps a manifest installation key to the newer registry
// version it can upgrade to. Only drifted installations appear in the map. The
// registry is read from the local cache only (never fetched); if no cached
// registry is available the map is nil and callers show no drift.
func availableUpdates(app *App, m *manifest.Manifest) map[string]string {
	ix := app.cachedSkillIndex()
	out := map[string]string{}
	for _, inst := range m.Installations {
		// Compare each install against its ORIGIN registry (falling back to any),
		// so a skill from registry A isn't mismarked against registry B.
		rs, found := ix.find(inst.RegistryName, inst.Skill)
		if !found {
			continue
		}
		// Drift signals mirror install.PlanUpdates: version or per-skill
		// DirSHA (RegistryRef) differs. The repo-wide Source.SHA is ignored.
		if inst.Version != rs.Version || inst.RegistryRef != rs.DirSHA {
			out[installKey(inst)] = rs.Version
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// installKey uniquely identifies one manifest installation row.
func installKey(inst manifest.Installation) string {
	return inst.Skill + "\x00" + inst.Platform + "\x00" + inst.Scope
}

// listInstallation is a manifest installation enriched with the newer registry
// version, if any. The embedded Installation's fields are promoted so the JSON
// shape stays backward-compatible; available_update is purely additive.
type listInstallation struct {
	manifest.Installation
	AvailableUpdate string `json:"available_update,omitempty"`
}

// listView is the --json document for `list`: identical to the manifest plus
// the additive per-row available_update field.
type listView struct {
	SchemaVersion int                `json:"schema_version"`
	Installations []listInstallation `json:"installations"`
}

func buildListView(m *manifest.Manifest, avail map[string]string) listView {
	v := listView{SchemaVersion: m.SchemaVersion}
	v.Installations = make([]listInstallation, 0, len(m.Installations))
	for _, inst := range m.Installations {
		v.Installations = append(v.Installations, listInstallation{
			Installation:    inst,
			AvailableUpdate: avail[installKey(inst)],
		})
	}
	return v
}

// renderListTable prints installs as a bordered table using the shared theme.
// Used in non-TTY / piped paths. "Source" is the canonical humblskills store
// every row's Path symlinks to — the source of truth — so scripted/piped
// output reflects the real install location, not just the platform copy.
func renderListTable(app *App, installs []manifest.Installation, avail map[string]string) {
	th := app.UI.Theme()
	app.UI.Header("list")

	// Show a Registry column only when installs actually span more than one
	// registry, so single-registry output stays uncluttered.
	regs := map[string]struct{}{}
	for _, inst := range installs {
		regs[inst.RegistryName] = struct{}{}
	}
	showRegistry := len(regs) > 1

	outdated := 0
	rows := make([][]string, 0, len(installs))
	for _, inst := range installs {
		source := inst.StorePath
		if source == "" {
			source = "-"
		}
		// When a newer registry version is known, show the transition
		// (v1.0.0 → v1.0.2) right in the Version cell instead of adding a
		// whole extra column to an already-wide table.
		version := th.Version.Render("v" + inst.Version)
		if to := avail[installKey(inst)]; to != "" && to != inst.Version {
			outdated++
			version = th.DotWarn.Render("↑ ") + th.Version.Render("v"+inst.Version) +
				th.Detail.Render(" → ") + th.DotWarn.Render("v"+to)
		}
		row := []string{th.Name.Render(inst.Skill), version}
		if showRegistry {
			row = append(row, th.Category.Render(registryDisplayName(inst.RegistryName)))
		}
		row = append(row,
			th.Platform.Render(inst.Platform),
			th.Label.Render(inst.Scope),
			th.Detail.Render(inst.Path),
			th.Detail.Render(source),
		)
		rows = append(rows, row)
	}
	headers := []string{"Skill", "Version"}
	if showRegistry {
		headers = append(headers, "Registry")
	}
	headers = append(headers, "Platform", "Scope", "Path", "Source")
	tbl := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(th.RuleLine).
		Headers(headers...).
		Rows(rows...).
		StyleFunc(func(row, _ int) lipgloss.Style {
			if row == table.HeaderRow {
				return th.Label.Padding(0, 1).Bold(true)
			}
			return lipgloss.NewStyle().Padding(0, 1)
		})
	fmt.Fprintln(app.UI.Out(), tbl.Render())
	fmt.Fprintln(app.UI.Out(), "  "+th.Crumb.Render(fmt.Sprintf(
		"%d install%s total", len(installs), textutil.Plural(len(installs)))))
	if outdated > 0 {
		app.UI.Warn("%d install%s can be updated - run 'humblskills update'",
			outdated, textutil.Plural(outdated))
	}
}

// runListTUI opens the shared two-pane browser in installed-only mode so list
// looks identical to search, differing only in which skills are shown and
// which actions are wired up.
func runListTUI(app *App, m *manifest.Manifest, fromDashboard bool) error {
	// Resolve each installed skill against its ORIGIN registry (across all
	// configured registries) so drift + metadata are correct and the browser
	// can group installs by source registry. Unreachable registries fall back
	// to manifest versions.
	ix := indexLoadedRegistries(app.loadRegistries())

	installedSkills := uniqueSkillsFromManifest(m)
	skills := make([]registry.Skill, 0, len(installedSkills))
	for _, name := range installedSkills {
		inst := m.FindAll(name)
		if len(inst) == 0 {
			continue
		}
		regName := inst[0].RegistryName
		sk, ok := ix.find(regName, name)
		if !ok {
			// No registry has it — synthesise from the manifest entry.
			sk = registry.Skill{Name: name, Version: inst[0].Version}
		}
		// Tag with the true origin from the manifest so grouping reflects where
		// it was installed from (not a fallback registry that happens to list it).
		sk.Registry = regName
		skills = append(skills, sk)
	}

	items := buildSkillItems(skills, m)

	skill, action, err := runSkillBrowser(app, "Installed", items, modeInstalledOnly, "no skills installed", fromDashboard)
	if err != nil {
		return err
	}
	switch action {
	case "update":
		return runUpdate(app, []string{skill}, updateFlags{})
	case "uninstall":
		return runUninstall(app, skill)
	}
	return nil
}

func uniqueSkillsFromManifest(m *manifest.Manifest) []string {
	seen := map[string]bool{}
	out := make([]string, 0)
	for _, inst := range m.Installations {
		if !seen[inst.Skill] {
			seen[inst.Skill] = true
			out = append(out, inst.Skill)
		}
	}
	sort.Strings(out)
	return out
}

func findRegistrySkill(reg *registry.Registry, name string) *registry.Skill {
	if reg == nil {
		return nil
	}
	for i := range reg.Skills {
		if reg.Skills[i].Name == name {
			return &reg.Skills[i]
		}
	}
	return nil
}
