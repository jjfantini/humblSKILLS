package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// RunProfileEditor opens a two-pane TUI for editing the user profile. The
// left pane lists settings (default platforms, default scope); pressing
// enter on a row opens an inline huh form to toggle/change that setting's
// value. Returns the edited profile and whether any change was made.
func RunProfileEditor(theme *ui.Theme, adapterList []adapters.Adapter, p *profile.Profile) (*profile.Profile, bool, error) {
	if p == nil {
		p = &profile.Profile{SchemaVersion: profile.SchemaVersion}
	}
	cur := *p
	changed := false

	for {
		items := []Item{
			profilePlatformsItem{profile: &cur, adapters: adapterList},
			profileScopeItem{profile: &cur},
		}
		res, err := RunListDetail(Config{
			Theme:      theme,
			Section:    "Profile",
			Items:      items,
			LeftTitle:  "SETTINGS",
			RightTitle: "VALUE",
			Actions: []ActionSpec{
				{Key: "e", Label: "edit", Action: "edit"},
			},
		})
		if err != nil {
			return &cur, changed, err
		}
		if res.Quit || res.Item == nil {
			return &cur, changed, nil
		}
		if res.Action != "edit" {
			continue
		}

		switch it := res.Item.(type) {
		case profilePlatformsItem:
			plats, ok, err := editProfilePlatforms(theme, adapterList, cur.DefaultPlatforms)
			if err != nil {
				return &cur, changed, err
			}
			if ok {
				cur.DefaultPlatforms = plats
				changed = true
			}
			_ = it
		case profileScopeItem:
			scope, ok, err := editProfileScope(theme, cur.DefaultScope)
			if err != nil {
				return &cur, changed, err
			}
			if ok {
				cur.DefaultScope = scope
				changed = true
			}
		}
	}
}

// editProfilePlatforms pops a huh MultiSelect over every adapter, returning
// the new platform list and ok=false if the user cancelled.
func editProfilePlatforms(theme *ui.Theme, adapterList []adapters.Adapter, current []string) ([]string, bool, error) {
	selected := append([]string(nil), current...)
	opts := make([]huh.Option[string], 0, len(adapterList))
	for _, a := range adapterList {
		opts = append(opts, huh.NewOption(a.Name, a.Name))
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Default platforms").
				Description("Leave empty to use every detected platform.").
				Options(opts...).
				Value(&selected),
		),
	).WithTheme(ui.HuhTheme(theme))
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return current, false, nil
		}
		return current, false, err
	}
	return selected, true, nil
}

// editProfileScope pops a huh Select with the three scope options.
func editProfileScope(theme *ui.Theme, current string) (string, bool, error) {
	scope := current
	opts := []huh.Option[string]{
		huh.NewOption("adapter default", ""),
		huh.NewOption("user", "user"),
		huh.NewOption("project", "project"),
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Default scope").
				Options(opts...).
				Value(&scope),
		),
	).WithTheme(ui.HuhTheme(theme))
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return current, false, nil
		}
		return current, false, err
	}
	return scope, true, nil
}

// --- list items -------------------------------------------------------------

type profilePlatformsItem struct {
	profile  *profile.Profile
	adapters []adapters.Adapter
}

func (p profilePlatformsItem) Key() string { return "default platforms" }
func (p profilePlatformsItem) FilterValue() string {
	return "platforms " + strings.Join(p.profile.DefaultPlatforms, " ")
}

func (p profilePlatformsItem) NaturalWidth(th *ui.Theme) int {
	badge := Badge(th, BadgeGhost, p.badgeText())
	return rowNaturalWidthProfile("default platforms", lipgloss.Width(badge))
}

func (p profilePlatformsItem) Row(th *ui.Theme, width int, selected bool) string {
	dot := th.DotOK.Render("●")
	if len(p.profile.DefaultPlatforms) == 0 {
		dot = th.DotNo.Render("●")
	}
	name := rowNameProfile(th, "default platforms", selected)
	badge := Badge(th, BadgeGhost, p.badgeText())
	return rowWithTrailingBadgeProfile(dot+" "+name, badge, width)
}

func (p profilePlatformsItem) Detail(th *ui.Theme, width int) string {
	var sb strings.Builder
	sb.WriteString(th.DetailTitle.Render("default platforms") + "\n")
	sb.WriteString(th.DetailSub.Render("press enter to edit") + "\n\n")

	selected := map[string]bool{}
	for _, n := range p.profile.DefaultPlatforms {
		selected[n] = true
	}

	if len(p.profile.DefaultPlatforms) == 0 {
		sb.WriteString(th.Desc.Width(width).Render(
			"No defaults set — installs target every detected platform.") + "\n\n")
	}

	sb.WriteString(th.SectionTitle.Render("PLATFORMS") + "\n")
	for _, a := range p.adapters {
		mark := th.RowDim.Render("[ ]")
		name := th.RowUnselected.Render(a.Name)
		if selected[a.Name] {
			mark = th.Success.Render("[✓]")
			name = th.RowSelected.Render(a.Name)
		}
		sb.WriteString("  " + mark + "  " + name + "\n")
	}
	_ = width
	return sb.String()
}

func (p profilePlatformsItem) badgeText() string {
	if len(p.profile.DefaultPlatforms) == 0 {
		return "all detected"
	}
	return fmt.Sprintf("%d platform%s", len(p.profile.DefaultPlatforms), pluralProfile(len(p.profile.DefaultPlatforms)))
}

type profileScopeItem struct {
	profile *profile.Profile
}

func (s profileScopeItem) Key() string         { return "default scope" }
func (s profileScopeItem) FilterValue() string { return "scope " + s.profile.DefaultScope }

func (s profileScopeItem) NaturalWidth(th *ui.Theme) int {
	badge := Badge(th, BadgeGhost, s.badgeText())
	return rowNaturalWidthProfile("default scope", lipgloss.Width(badge))
}

func (s profileScopeItem) Row(th *ui.Theme, width int, selected bool) string {
	dot := th.DotOK.Render("●")
	if s.profile.DefaultScope == "" {
		dot = th.DotNo.Render("●")
	}
	name := rowNameProfile(th, "default scope", selected)
	badge := Badge(th, BadgeGhost, s.badgeText())
	return rowWithTrailingBadgeProfile(dot+" "+name, badge, width)
}

func (s profileScopeItem) Detail(th *ui.Theme, width int) string {
	var sb strings.Builder
	sb.WriteString(th.DetailTitle.Render("default scope") + "\n")
	sb.WriteString(th.DetailSub.Render("press enter to edit") + "\n\n")
	sb.WriteString(th.Desc.Width(width).Render(
		"Which target scope to use when installing. Adapter default falls back to each adapter's own default (usually user).") + "\n\n")

	sb.WriteString(th.SectionTitle.Render("OPTIONS") + "\n")
	for _, opt := range []struct {
		label string
		value string
	}{
		{"adapter default", ""},
		{"user", "user"},
		{"project", "project"},
	} {
		mark := th.RowDim.Render("[ ]")
		label := th.RowUnselected.Render(opt.label)
		if s.profile.DefaultScope == opt.value {
			mark = th.Success.Render("[✓]")
			label = th.RowSelected.Render(opt.label)
		}
		sb.WriteString("  " + mark + "  " + label + "\n")
	}
	_ = width
	return sb.String()
}

func (s profileScopeItem) badgeText() string {
	if s.profile.DefaultScope == "" {
		return "adapter default"
	}
	return s.profile.DefaultScope
}

// --- small helpers (kept local so this file is self-contained) --------------

func rowNaturalWidthProfile(label string, badgeWidth int) int {
	// 1 (dot) + 1 (space) + label + 2 (gap) + badge.
	return 1 + 1 + lipgloss.Width(label) + 2 + badgeWidth
}

func rowNameProfile(th *ui.Theme, name string, selected bool) string {
	if selected {
		return th.RowSelected.Render(name)
	}
	return th.RowUnselected.Render(name)
}

func rowWithTrailingBadgeProfile(label, badge string, width int) string {
	lw := lipgloss.Width(label)
	if width < 10 || width-lw < lipgloss.Width(badge)+1 {
		if lw >= width {
			return label
		}
		return label + strings.Repeat(" ", width-lw)
	}
	gap := width - lw - lipgloss.Width(badge)
	return label + strings.Repeat(" ", gap) + badge
}

func pluralProfile(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// --- per-install modal ------------------------------------------------------

// InstallModalResult is what the install platform picker returns.
type InstallModalResult struct {
	Platforms   []string
	Scope       string
	Confirmed   bool
	EditProfile bool
}

// RunInstallPlatformModal opens a huh form asking the user which detected
// platforms to install `skill` into, and at which scope. Default selections
// come from the profile. Returns Confirmed=false if the user cancelled, and
// EditProfile=true if they chose the "edit profile" action.
func RunInstallPlatformModal(
	theme *ui.Theme,
	skill string,
	adapterList []adapters.Adapter,
	detected map[string]bool,
	p *profile.Profile,
) (InstallModalResult, error) {
	if p == nil {
		p = &profile.Profile{}
	}

	platforms := make([]string, 0)
	if len(p.DefaultPlatforms) > 0 {
		for _, name := range p.DefaultPlatforms {
			if detected[name] {
				platforms = append(platforms, name)
			}
		}
	} else {
		for _, a := range adapterList {
			if detected[a.Name] {
				platforms = append(platforms, a.Name)
			}
		}
	}

	opts := make([]huh.Option[string], 0, len(adapterList))
	for _, a := range adapterList {
		label := a.Name
		if detected[a.Name] {
			label += "  (detected)"
		} else {
			label += "  (not detected)"
		}
		opts = append(opts, huh.NewOption(label, a.Name))
	}

	scope := p.DefaultScope
	scopeOpts := []huh.Option[string]{
		huh.NewOption("adapter default", ""),
		huh.NewOption("user", "user"),
		huh.NewOption("project", "project"),
	}

	action := "install"
	actionOpts := []huh.Option[string]{
		huh.NewOption("install", "install"),
		huh.NewOption("edit profile defaults", "profile"),
		huh.NewOption("cancel", "cancel"),
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Install "+skill+" to:").
				Options(opts...).
				Value(&platforms),
			huh.NewSelect[string]().
				Title("Scope").
				Options(scopeOpts...).
				Value(&scope),
			huh.NewSelect[string]().
				Title("Action").
				Options(actionOpts...).
				Value(&action),
		),
	).WithTheme(ui.HuhTheme(theme))

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return InstallModalResult{}, nil
		}
		return InstallModalResult{}, err
	}

	switch action {
	case "profile":
		return InstallModalResult{EditProfile: true}, nil
	case "cancel":
		return InstallModalResult{}, nil
	}

	return InstallModalResult{
		Platforms: platforms,
		Scope:     scope,
		Confirmed: true,
	}, nil
}
