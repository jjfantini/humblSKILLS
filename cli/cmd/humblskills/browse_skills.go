package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/textutil"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// skillsBrowseMode determines the verb shown in the footer and the result
// action returned when the user presses enter.
type skillsBrowseMode int

const (
	// modeSearch: every skill is selectable; enter triggers install.
	modeSearch skillsBrowseMode = iota
	// modeInstalledOnly: only installed skills are shown; enter triggers
	// update, `x` triggers uninstall.
	modeInstalledOnly
)

// skillItem adapts a registry.Skill (with optional installed-state overlay)
// to the tui.Item interface. Shared by search, install, list, uninstall.
//
// installs holds *every* manifest entry for this skill — one per
// (platform, scope) the engine wrote a symlink for — not just the first one
// found. That's what lets Detail() show the canonical humblskills store
// (source of truth) plus every platform it's symlinked to, instead of
// whichever platform happened to be inserted into the manifest first.
type skillItem struct {
	s        registry.Skill
	installs []manifest.Installation // empty if not installed on this machine
	outdated bool                    // any installed entry's version != registry version
}

// primary returns the first installed entry, or nil if none. Used where a
// single representative install is enough (e.g. the left-pane row dot).
func (s skillItem) primary() *manifest.Installation {
	if len(s.installs) == 0 {
		return nil
	}
	return &s.installs[0]
}

func (s skillItem) Key() string { return s.s.Name }
func (s skillItem) FilterValue() string {
	return strings.ToLower(s.s.Name + " " + strings.Join(s.s.Tags, " ") + " " + s.s.Description)
}

// NaturalWidth reports the row's display width: dot + space + name + 2-gap
// + version. Status is encoded entirely in the leading dot colour
// (green = installed, yellow = outdated, red = not installed), so no
// trailing badge is needed.
func (s skillItem) NaturalWidth(th *ui.Theme) int {
	_ = th
	versionW := lipgloss.Width("v" + s.s.Version)
	return 1 + 1 + lipgloss.Width(s.s.Name) + 2 + versionW
}

func (s skillItem) Row(th *ui.Theme, width int, selected bool) string {
	var dot string
	switch {
	case len(s.installs) == 0:
		dot = th.DotNo.Render("●")
	case s.outdated:
		dot = th.DotWarn.Render("●")
	default:
		dot = th.DotOK.Render("●")
	}

	name := rowName(th, s.s.Name, selected, true)
	version := th.Version.Render("v" + s.s.Version)

	row := dot + " " + name + "  " + version
	rw := lipgloss.Width(row)
	if rw >= width {
		return row
	}
	return row + strings.Repeat(" ", width-rw)
}

func (s skillItem) Detail(th *ui.Theme, width int) string {
	var sb strings.Builder
	primary := s.primary()
	sub := "v" + s.s.Version
	if primary != nil && primary.Version != s.s.Version {
		sub = "v" + primary.Version + " → v" + s.s.Version
	}
	sb.WriteString(th.DetailTitle.Render(s.s.Name) + "  " +
		th.DetailSub.Render(sub) + "\n\n")

	if s.s.Description != "" {
		desc := th.Desc.Width(width).Render(s.s.Description)
		sb.WriteString(desc + "\n\n")
	}

	if len(s.s.Tags) > 0 {
		chips := make([]string, 0, len(s.s.Tags))
		for _, t := range s.s.Tags {
			chips = append(chips, th.Tag.Render("#"+t))
		}
		sb.WriteString(kvRow(th, "tags", strings.Join(chips, "  ")))
	}
	if len(s.s.Platforms) > 0 {
		plats := make([]string, 0, len(s.s.Platforms))
		for _, p := range s.s.Platforms {
			plats = append(plats, th.Platform.Render(p))
		}
		sb.WriteString(kvRow(th, "target", strings.Join(plats, "  ")))
	}
	if len(s.s.Requires) > 0 {
		sb.WriteString(kvRow(th, "deps", th.KVValue.Render(strings.Join(s.s.Requires, ", "))))
	}

	if len(s.installs) > 0 {
		sb.WriteString(s.installedSection(th, width))
	}
	return sb.String()
}

// installedSection renders the "INSTALLED" block: the canonical humblskills
// store (source of truth — every platform below is a symlink to this one
// directory) followed by every platform/scope it's symlinked to. Falls back
// gracefully for legacy manifest entries written before store_path existed.
func (s skillItem) installedSection(th *ui.Theme, width int) string {
	var sb strings.Builder
	sb.WriteString("\n" + th.SectionTitle.Render("INSTALLED") + "\n")

	store := ""
	for _, inst := range s.installs {
		if inst.StorePath != "" {
			store = inst.StorePath
			break
		}
	}
	if store != "" {
		sb.WriteString(kvRow(th, "store", th.KVValue.Render(store)))
	}

	sb.WriteString("\n" + th.SectionTitle.Render("SYMLINKED PLATFORMS") + "\n")
	for i, inst := range s.installs {
		if i > 0 {
			sb.WriteString(tui.DashedRule(th, width) + "\n")
		}
		ver := "v" + inst.Version
		if inst.Version != s.s.Version {
			ver = "v" + inst.Version + " → v" + s.s.Version
		}
		sb.WriteString(kvRow(th, "platform", th.Platform.Render(inst.Platform)+"  "+th.KVValue.Render(ver)))
		sb.WriteString(kvRow(th, "scope", th.KVValue.Render(inst.Scope)))
		sb.WriteString(kvRow(th, "path", th.KVValue.Render(inst.Path)))
		if !inst.InstalledAt.IsZero() {
			sb.WriteString(kvRow(th, "at", th.KVValue.Render(
				inst.InstalledAt.Format("2006-01-02 15:04"))))
		}
	}
	return sb.String()
}

// buildSkillItems joins a registry listing with the install manifest so the
// returned items know every (platform, scope) they're installed onto, and
// whether any of those have drifted from the registry version.
func buildSkillItems(skills []registry.Skill, m *manifest.Manifest) []skillItem {
	installed := map[string][]manifest.Installation{}
	if m != nil {
		for _, inst := range m.Installations {
			installed[inst.Skill] = append(installed[inst.Skill], inst)
		}
		for name, insts := range installed {
			sort.SliceStable(insts, func(i, j int) bool {
				if insts[i].Platform != insts[j].Platform {
					return insts[i].Platform < insts[j].Platform
				}
				return insts[i].Scope < insts[j].Scope
			})
			installed[name] = insts
		}
	}
	items := make([]skillItem, 0, len(skills))
	for _, s := range skills {
		it := skillItem{s: s}
		if insts, ok := installed[s.Name]; ok {
			it.installs = insts
			for _, inst := range insts {
				if inst.Version != s.Version {
					it.outdated = true
					break
				}
			}
		}
		items = append(items, it)
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].s.Name < items[j].s.Name
	})
	return items
}

// runSkillBrowser opens the shared two-pane picker over skills and routes the
// user's choice through the right subcommand. Returns (skill, action) where
// action is one of "install", "update", "uninstall", or "" (user quit).
//
// Pressing `p` opens the profile editor inline and re-enters the picker so
// every surface that uses this browser gets the same footer shortcut.
func runSkillBrowser(app *App, section string, skills []skillItem, mode skillsBrowseMode, emptyMsg string, fromDashboard bool) (string, string, error) {
	if len(skills) == 0 {
		app.UI.Info(emptyMsg)
		return "", "", nil
	}
	items := make([]tui.Item, 0, len(skills))
	for _, s := range skills {
		items = append(items, s)
	}

	var actions []tui.ActionSpec
	switch mode {
	case modeSearch:
		actions = []tui.ActionSpec{
			{Key: "i", Label: "install", Action: "install"},
			{Key: "p", Label: "profile", Action: "profile"},
		}
	case modeInstalledOnly:
		actions = []tui.ActionSpec{
			{Key: "u", Label: "update", Action: "update"},
			{Key: "x", Label: "uninstall", Action: "uninstall"},
			{Key: "p", Label: "profile", Action: "profile"},
		}
	}

	installedCount, outdatedCount := 0, 0
	for _, s := range skills {
		if len(s.installs) > 0 {
			installedCount++
		}
		if s.outdated {
			outdatedCount++
		}
	}
	localMeta := func(items []tui.Item, _ int) string {
		parts := []string{fmt.Sprintf("%d skill%s", len(items), textutil.Plural(len(items)))}
		if installedCount > 0 {
			parts = append(parts, fmt.Sprintf("%d installed", installedCount))
		}
		if outdatedCount > 0 {
			parts = append(parts, fmt.Sprintf("%d outdated", outdatedCount))
		}
		return strings.Join(parts, " · ")
	}
	// When launched from the dashboard, mirror the shared status line
	// ("● healthy · N platforms · M skills"); otherwise show the command-local
	// counts so direct `humblskills install`/`search` still feel informative.
	meta := localMeta
	if app.Nav.Crumb != "" {
		status := app.Nav.Status
		theme := app.UI.Theme()
		meta = func(_ []tui.Item, _ int) string {
			return tui.RenderStatusMeta(theme, status)
		}
	}

	leftTitle := "SKILLS"
	if mode == modeInstalledOnly {
		leftTitle = "INSTALLED"
	}

	for {
		cfg := tui.Config{
			Theme:      app.UI.Theme(),
			Version:    resolveVersion().Version,
			Section:    app.headerSection(section),
			Meta:       meta,
			Items:      items,
			LeftTitle:  leftTitle,
			RightTitle: "DETAIL",
			Actions:    actions,
			EmptyMsg:   emptyMsg,
		}
		if fromDashboard {
			cfg.BackKey = "esc"
			cfg.BackLabel = "back"
		}
		res, err := tui.RunListDetail(cfg)
		if err != nil {
			return "", "", err
		}
		if res.Quit || res.Item == nil {
			return "", "", nil
		}
		if res.Action == "profile" {
			if err := runProfileEditor(app); err != nil {
				return "", "", err
			}
			continue
		}
		it, ok := res.Item.(skillItem)
		if !ok {
			return "", "", nil
		}
		return it.s.Name, res.Action, nil
	}
}
