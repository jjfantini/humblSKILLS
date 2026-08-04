package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/v2/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/textutil"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/tui"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/ui"
)

const uncategorizedLabel = "uncategorized"

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
	s          registry.Skill
	installs   []manifest.Installation // empty if not installed on this machine
	outdated   bool                    // any installed entry's version != registry version
	parentKeys []string                // collapse keys of ancestor category/role headers
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
	return strings.ToLower(s.s.Name + " " + s.s.Category + " " + s.s.Role + " " + strings.Join(s.s.Tags, " ") + " " + s.s.Description)
}
func (s skillItem) ParentCollapseKeys() []string { return s.parentKeys }

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
	if width < 20 {
		width = 20
	}
	var sb strings.Builder
	primary := s.primary()
	sub := "v" + s.s.Version
	if primary != nil && primary.Version != s.s.Version {
		sub = "v" + primary.Version + " → v" + s.s.Version
	}
	title := th.DetailTitle.Render(s.s.Name) + "  " + th.DetailSub.Render(sub)
	sb.WriteString(ansiWrap(title, width) + "\n\n")

	if s.s.Description != "" {
		desc := th.Desc.Width(width).Render(s.s.Description)
		// Hard-wrap any leftover long tokens lipgloss left intact.
		var descOut []string
		for _, line := range strings.Split(desc, "\n") {
			descOut = append(descOut, strings.Split(ansiWrap(line, width), "\n")...)
		}
		sb.WriteString(strings.Join(descOut, "\n") + "\n\n")
	}

	if s.s.Registry != "" {
		sb.WriteString(kvRowWidth(th, "registry", th.KVValue.Render(s.s.Registry), width))
	}
	if s.s.Category != "" {
		sb.WriteString(kvRowWidth(th, "category", th.Category.Render(s.s.Category), width))
	}
	if s.s.Role != "" {
		sb.WriteString(kvRowWidth(th, "role", th.Platform.Render(s.s.Role), width))
	}
	if len(s.s.Tags) > 0 {
		chips := make([]string, 0, len(s.s.Tags))
		for _, t := range s.s.Tags {
			chips = append(chips, th.Tag.Render("#"+t))
		}
		sb.WriteString(kvRowWidth(th, "tags", strings.Join(chips, "  "), width))
	}
	if len(s.s.Platforms) > 0 {
		plats := make([]string, 0, len(s.s.Platforms))
		for _, p := range s.s.Platforms {
			plats = append(plats, th.Platform.Render(p))
		}
		sb.WriteString(kvRowWidth(th, "target", strings.Join(plats, "  "), width))
	}
	if len(s.s.Requires) > 0 {
		sb.WriteString(kvRowWidth(th, "deps", th.KVValue.Render(strings.Join(s.s.Requires, ", ")), width))
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
		sb.WriteString(kvRowWidth(th, "store", th.KVValue.Render(store), width))
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
		sb.WriteString(kvRowWidth(th, "platform", th.Platform.Render(inst.Platform)+"  "+th.KVValue.Render(ver), width))
		sb.WriteString(kvRowWidth(th, "scope", th.KVValue.Render(inst.Scope), width))
		sb.WriteString(kvRowWidth(th, "path", th.KVValue.Render(inst.Path), width))
		if !inst.InstalledAt.IsZero() {
			sb.WriteString(kvRowWidth(th, "at", th.KVValue.Render(
				inst.InstalledAt.Format("2006-01-02 15:04")), width))
		}
	}
	return sb.String()
}

// buildSkillItems joins a registry listing with the install manifest so the
// returned items know every (platform, scope) they're installed onto, and
// whether any of those have drifted from the registry version.
// groupByCategory controls sort order (must match aggregateSkills / buildSkillTree).
func buildSkillItems(skills []registry.Skill, m *manifest.Manifest, groupByCategory bool) []skillItem {
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
		return lessSkill(items[i].s, items[j].s, groupByCategory)
	})
	return items
}

// lessSkill orders skills for the TUI tree. With groupByCategory:
// registry → category → role → name. Without: registry → role → name.
func lessSkill(a, b registry.Skill, groupByCategory bool) bool {
	if a.Registry != b.Registry {
		return a.Registry < b.Registry
	}
	if groupByCategory {
		ca, cb := skillCategory(a), skillCategory(b)
		if ca != cb {
			return ca < cb
		}
	}
	if a.Role != b.Role {
		return a.Role < b.Role
	}
	return a.Name < b.Name
}

func skillCategory(s registry.Skill) string {
	if strings.TrimSpace(s.Category) == "" {
		return uncategorizedLabel
	}
	return s.Category
}

// sectionKind distinguishes registry dividers from collapsible category/role
// headers in the skills list.
type sectionKind int

const (
	sectionRegistry sectionKind = iota
	sectionCategory
	sectionRole
)

// sectionHeaderItem is a list section row. Registry headers are non-focusable
// full-width dividers (HeaderItem). Category and role headers are focusable
// CollapsibleItems toggled with Enter.
type sectionHeaderItem struct {
	kind       sectionKind
	name       string
	registry   string
	category   string   // set for role headers (owning category)
	parentKeys []string // ancestor collapse keys
}

func (h sectionHeaderItem) FilterValue() string { return "" }

func (h sectionHeaderItem) Key() string {
	switch h.kind {
	case sectionRegistry:
		return "__header__:reg:" + h.registry
	case sectionCategory:
		return "__header__:cat:" + h.registry + ":" + h.name
	case sectionRole:
		return "__header__:role:" + h.registry + ":" + h.category + ":" + h.name
	}
	return "__header__:" + h.name
}

func (h sectionHeaderItem) IsHeader() bool { return h.kind == sectionRegistry }

func (h sectionHeaderItem) IsCollapsible() bool {
	return h.kind == sectionCategory || h.kind == sectionRole
}

func (h sectionHeaderItem) CollapseKey() string {
	return h.Key()
}

func (h sectionHeaderItem) ParentCollapseKeys() []string { return h.parentKeys }

func (h sectionHeaderItem) NaturalWidth(th *ui.Theme) int {
	return lipgloss.Width(h.renderLabel(th, false))
}

func (h sectionHeaderItem) renderLabel(th *ui.Theme, collapsed bool) string {
	switch h.kind {
	case sectionRegistry:
		return th.SectionTitle.Render("── " + h.name + " ")
	case sectionCategory:
		caret := "▾"
		if collapsed {
			caret = "▸"
		}
		return th.SectionTitle.Render(caret + " " + h.name)
	case sectionRole:
		caret := "▾"
		if collapsed {
			caret = "▸"
		}
		return th.Meta.Render("  " + caret + " " + h.name)
	}
	return h.name
}

func (h sectionHeaderItem) Row(th *ui.Theme, width int, selected bool) string {
	return h.RowCollapsed(th, width, selected, false)
}

func (h sectionHeaderItem) RowCollapsed(th *ui.Theme, width int, selected, collapsed bool) string {
	if h.kind == sectionRegistry {
		// Full-width registry bar: "── name ────────"
		base := "── " + h.name + " "
		pad := width - lipgloss.Width(base)
		if pad < 2 {
			pad = 2
		}
		return th.SectionTitle.Render(base + strings.Repeat("─", pad))
	}
	label := h.renderLabel(th, collapsed)
	if selected {
		// Keep caret + name; selection bar is drawn by the list model.
		_ = selected
	}
	rw := lipgloss.Width(label)
	if rw >= width {
		return label
	}
	return label + strings.Repeat(" ", width-rw)
}

func (h sectionHeaderItem) Detail(th *ui.Theme, width int) string {
	if width < 20 {
		width = 20
	}
	var sb strings.Builder
	switch h.kind {
	case sectionRegistry:
		sb.WriteString(th.DetailTitle.Render(h.name) + "\n\n")
		sb.WriteString(th.Desc.Width(width).Render("Registry section. Skills below are grouped from this source.") + "\n")
	case sectionCategory:
		sb.WriteString(th.DetailTitle.Render(h.name) + "\n\n")
		sb.WriteString(th.Desc.Width(width).Render("Category group. Press enter to expand or collapse skills in this category.") + "\n")
		sb.WriteString(kvRowWidth(th, "registry", th.KVValue.Render(registryDisplayName(h.registry)), width))
	case sectionRole:
		sb.WriteString(th.DetailTitle.Render(h.name) + "\n\n")
		sb.WriteString(th.Desc.Width(width).Render("Role group. Press enter to expand or collapse skills for this role.") + "\n")
		sb.WriteString(kvRowWidth(th, "registry", th.KVValue.Render(registryDisplayName(h.registry)), width))
		sb.WriteString(kvRowWidth(th, "category", th.Category.Render(h.category), width))
	}
	return sb.String()
}

// buildSkillTree turns a sorted skillItem list into tui.Items with section
// headers. When groupByCategory is true: Registry → Category → Role → Skills.
// When false (legacy): Registry → Role → Skills.
// Registry headers appear only when multiRegistry is true.
func buildSkillTree(skills []skillItem, multiRegistry, groupByCategory bool) []tui.Item {
	if groupByCategory {
		return buildSkillTreeByCategory(skills, multiRegistry)
	}
	return buildSkillTreeByRole(skills, multiRegistry)
}

func buildSkillTreeByCategory(skills []skillItem, multiRegistry bool) []tui.Item {
	items := make([]tui.Item, 0, len(skills)+8)
	prevReg, prevCat, prevRole := "", "", ""
	started := false
	for _, s := range skills {
		reg := s.s.Registry
		cat := skillCategory(s.s)
		role := s.s.Role
		regChanged := !started || reg != prevReg
		catChanged := regChanged || cat != prevCat
		roleChanged := catChanged || role != prevRole

		if multiRegistry && regChanged {
			items = append(items, sectionHeaderItem{
				kind:     sectionRegistry,
				name:     registryDisplayName(reg),
				registry: reg,
			})
		}
		var catKey string
		if catChanged {
			catHeader := sectionHeaderItem{
				kind:     sectionCategory,
				name:     cat,
				registry: reg,
			}
			catKey = catHeader.CollapseKey()
			items = append(items, catHeader)
		} else {
			catKey = sectionHeaderItem{kind: sectionCategory, name: cat, registry: reg}.CollapseKey()
		}

		parentKeys := []string{catKey}
		if role != "" && roleChanged {
			roleHeader := sectionHeaderItem{
				kind:       sectionRole,
				name:       role,
				registry:   reg,
				category:   cat,
				parentKeys: []string{catKey},
			}
			items = append(items, roleHeader)
			parentKeys = append(parentKeys, roleHeader.CollapseKey())
		} else if role != "" {
			roleKey := sectionHeaderItem{
				kind: sectionRole, name: role, registry: reg, category: cat,
			}.CollapseKey()
			parentKeys = append(parentKeys, roleKey)
		}

		s.parentKeys = parentKeys
		items = append(items, s)
		prevReg, prevCat, prevRole, started = reg, cat, role, true
	}
	return items
}

// buildSkillTreeByRole is the legacy layout: Registry → Role → Skills.
func buildSkillTreeByRole(skills []skillItem, multiRegistry bool) []tui.Item {
	hasRoles := false
	for _, s := range skills {
		if s.s.Role != "" {
			hasRoles = true
			break
		}
	}
	items := make([]tui.Item, 0, len(skills)+4)
	prevReg, prevRole := "", ""
	started := false
	for _, s := range skills {
		regChanged := !started || s.s.Registry != prevReg
		if multiRegistry && regChanged {
			items = append(items, sectionHeaderItem{
				kind:     sectionRegistry,
				name:     registryDisplayName(s.s.Registry),
				registry: s.s.Registry,
			})
		}
		var parentKeys []string
		if hasRoles && s.s.Role != "" && (regChanged || s.s.Role != prevRole) {
			roleHeader := sectionHeaderItem{
				kind:     sectionRole,
				name:     s.s.Role,
				registry: s.s.Registry,
				category: "",
			}
			items = append(items, roleHeader)
			parentKeys = []string{roleHeader.CollapseKey()}
		} else if hasRoles && s.s.Role != "" {
			roleKey := sectionHeaderItem{
				kind: sectionRole, name: s.s.Role, registry: s.s.Registry,
			}.CollapseKey()
			parentKeys = []string{roleKey}
		}
		s.parentKeys = parentKeys
		items = append(items, s)
		prevReg, prevRole, started = s.s.Registry, s.s.Role, true
	}
	return items
}

// interleaveRegistryHeaders is kept as a thin alias for tests / callers that
// still use the legacy name. It builds the category tree (default-on behaviour).
func interleaveRegistryHeaders(skills []skillItem, grouped bool) []tui.Item {
	return buildSkillTree(skills, grouped, true)
}

func registryDisplayName(name string) string {
	if name == "" {
		return "default"
	}
	return name
}

// countSkills counts the real skill rows in a tui.Item list (skips headers
// and collapsible section rows).
func countSkills(items []tui.Item) int {
	n := 0
	for _, it := range items {
		if _, ok := it.(skillItem); ok {
			n++
		}
	}
	return n
}

// distinctRegistries counts the distinct source-registry names across skills.
func distinctRegistries(skills []skillItem) int {
	seen := map[string]struct{}{}
	for _, s := range skills {
		seen[s.s.Registry] = struct{}{}
	}
	return len(seen)
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
		n := countSkills(items)
		parts := []string{fmt.Sprintf("%d skill%s", n, textutil.Plural(n))}
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
		groupBy := app.resolvedGroupByCategory()
		// Re-sort + rebuild tree each loop so a profile toggle takes effect
		// immediately after the inline profile editor returns.
		sorted := append([]skillItem(nil), skills...)
		sort.SliceStable(sorted, func(i, j int) bool {
			return lessSkill(sorted[i].s, sorted[j].s, groupBy)
		})
		items := buildSkillTree(sorted, distinctRegistries(sorted) > 1, groupBy)

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

// resolvedGroupByCategory loads the profile toggle; defaults to on when the
// profile is missing or unreadable so the TUI still groups by category.
func (app *App) resolvedGroupByCategory() bool {
	p, err := profile.Load(app.Config.ProfilePath)
	if err != nil || p == nil {
		return true
	}
	return p.ResolvedGroupByCategory()
}
