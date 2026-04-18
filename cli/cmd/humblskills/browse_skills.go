package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
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
type skillItem struct {
	s         registry.Skill
	installed *manifest.Installation // nil if not installed on this machine
	outdated  bool                   // registry version > installed version
}

func (s skillItem) Key() string { return s.s.Name }
func (s skillItem) FilterValue() string {
	return strings.ToLower(s.s.Name + " " + strings.Join(s.s.Tags, " ") + " " + s.s.Description)
}

func (s skillItem) Row(th *ui.Theme, width int, selected bool) string {
	var dot string
	if s.installed != nil {
		if s.outdated {
			dot = th.DotWarn.Render("●")
		} else {
			dot = th.DotOK.Render("●")
		}
	} else {
		dot = th.DotNo.Render("●")
	}

	name := rowName(th, s.s.Name, selected, true)
	version := th.Version.Render("v" + s.s.Version)

	left := dot + " " + name + "  " + version

	var badge string
	if s.outdated {
		badge = tui.Badge(th, tui.BadgeRO, "outdated")
	} else if s.installed != nil {
		badge = tui.Badge(th, tui.BadgeDetected, "installed")
	}
	if badge == "" {
		return left
	}
	return rowWithTrailingBadge(left, badge, width)
}

func (s skillItem) Detail(th *ui.Theme, width int) string {
	var sb strings.Builder
	sub := "v" + s.s.Version
	if s.installed != nil && s.installed.Version != s.s.Version {
		sub = "v" + s.installed.Version + " → v" + s.s.Version
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

	if s.installed != nil {
		sb.WriteString("\n" + th.SectionTitle.Render("INSTALLED") + "\n")
		sb.WriteString(kvRow(th, "version", th.KVValue.Render("v"+s.installed.Version)))
		sb.WriteString(kvRow(th, "platform", th.KVValue.Render(s.installed.Platform)))
		sb.WriteString(kvRow(th, "scope", th.KVValue.Render(s.installed.Scope)))
		sb.WriteString(kvRow(th, "path", th.KVValue.Render(s.installed.Path)))
		if !s.installed.InstalledAt.IsZero() {
			sb.WriteString(kvRow(th, "at", th.KVValue.Render(
				s.installed.InstalledAt.Format("2006-01-02 15:04"))))
		}
	}
	return sb.String()
}

// buildSkillItems joins a registry listing with the install manifest so the
// returned items know whether they're installed and/or drifted.
func buildSkillItems(skills []registry.Skill, m *manifest.Manifest) []skillItem {
	installed := map[string]*manifest.Installation{}
	if m != nil {
		for i := range m.Installations {
			it := &m.Installations[i]
			if _, seen := installed[it.Skill]; !seen {
				installed[it.Skill] = it
			}
		}
	}
	items := make([]skillItem, 0, len(skills))
	for _, s := range skills {
		it := skillItem{s: s}
		if inst, ok := installed[s.Name]; ok {
			it.installed = inst
			if inst.Version != s.Version {
				it.outdated = true
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
func runSkillBrowser(app *App, section string, skills []skillItem, mode skillsBrowseMode, emptyMsg string) (string, string, error) {
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
		}
	case modeInstalledOnly:
		actions = []tui.ActionSpec{
			{Key: "u", Label: "update", Action: "update"},
			{Key: "x", Label: "uninstall", Action: "uninstall"},
		}
	}

	installedCount, outdatedCount := 0, 0
	for _, s := range skills {
		if s.installed != nil {
			installedCount++
		}
		if s.outdated {
			outdatedCount++
		}
	}
	meta := func(items []tui.Item, _ int) string {
		parts := []string{fmt.Sprintf("%d skill%s", len(items), plural(len(items)))}
		if installedCount > 0 {
			parts = append(parts, fmt.Sprintf("%d installed", installedCount))
		}
		if outdatedCount > 0 {
			parts = append(parts, fmt.Sprintf("%d outdated", outdatedCount))
		}
		return strings.Join(parts, " · ")
	}

	leftTitle := "SKILLS"
	if mode == modeInstalledOnly {
		leftTitle = "INSTALLED"
	}

	res, err := tui.RunListDetail(tui.Config{
		Theme:      app.UI.Theme(),
		Version:    resolveVersion().Version,
		Section:    section,
		Meta:       meta,
		Items:      items,
		LeftTitle:  leftTitle,
		RightTitle: "DETAIL",
		Actions:    actions,
		EmptyMsg:   emptyMsg,
	})
	if err != nil {
		return "", "", err
	}
	if res.Quit || res.Item == nil {
		return "", "", nil
	}
	it, ok := res.Item.(skillItem)
	if !ok {
		return "", "", nil
	}
	return it.s.Name, res.Action, nil
}
