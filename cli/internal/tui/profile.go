package tui

import (
	"errors"

	"github.com/charmbracelet/huh"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// RunProfileEditor opens a huh form to edit the user's profile (default
// platforms + default scope). Returns (updated profile, saved, err). `saved`
// is false if the user cancelled.
func RunProfileEditor(theme *ui.Theme, adapterList []adapters.Adapter, p *profile.Profile) (*profile.Profile, bool, error) {
	if p == nil {
		p = &profile.Profile{}
	}

	platforms := make([]string, len(p.DefaultPlatforms))
	copy(platforms, p.DefaultPlatforms)
	scope := p.DefaultScope

	opts := make([]huh.Option[string], 0, len(adapterList))
	for _, a := range adapterList {
		opts = append(opts, huh.NewOption(a.Name, a.Name))
	}

	scopeOpts := []huh.Option[string]{
		huh.NewOption("adapter default", ""),
		huh.NewOption("user", "user"),
		huh.NewOption("project", "project"),
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Default platforms").
				Description("Leave empty to use every detected platform.").
				Options(opts...).
				Value(&platforms),
			huh.NewSelect[string]().
				Title("Default scope").
				Options(scopeOpts...).
				Value(&scope),
		),
	).WithTheme(ui.HuhTheme(theme))

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return p, false, nil
		}
		return p, false, err
	}

	out := &profile.Profile{
		SchemaVersion:    profile.SchemaVersion,
		DefaultPlatforms: platforms,
		DefaultScope:     scope,
	}
	return out, true, nil
}

// InstallModalResult is what the install platform picker returns.
type InstallModalResult struct {
	Platforms []string
	Scope     string
	Confirmed bool
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
