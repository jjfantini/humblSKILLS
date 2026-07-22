package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/secrets"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// registryMgrItem adapts a configured registry to the shared list widget.
type registryMgrItem struct {
	name string
	url  string
}

func (r registryMgrItem) Key() string { return r.name }
func (r registryMgrItem) FilterValue() string {
	return strings.ToLower(r.name + " " + r.url)
}
func (r registryMgrItem) NaturalWidth(th *ui.Theme) int {
	return 1 + 1 + lipgloss.Width(r.name)
}

func (r registryMgrItem) Row(th *ui.Theme, width int, selected bool) string {
	dot := th.DotNo.Render("●")
	if registryHasToken(r.name) {
		dot = th.DotOK.Render("●")
	}
	row := dot + " " + rowName(th, r.name, selected, true)
	if rw := lipgloss.Width(row); rw < width {
		row += strings.Repeat(" ", width-rw)
	}
	return row
}

func (r registryMgrItem) Detail(th *ui.Theme, width int) string {
	var sb strings.Builder
	sb.WriteString(th.DetailTitle.Render(r.name) + "\n\n")
	sb.WriteString(kvRow(th, "url", th.KVValue.Render(r.url)))
	sb.WriteString(kvRow(th, "token", th.KVValue.Render(registryTokenLabel(r.name))))
	return sb.String()
}

// runRegistryManager opens the interactive registry manager: a list of
// configured registries with add / rename / login / logout / remove / refresh
// actions. Falls back to a static list on non-TTY.
func runRegistryManager(app *App) error {
	if !tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes) {
		return runRegistryList(app)
	}

	fromDashboard := app.Nav.Crumb != ""
	for {
		p, err := profile.Load(app.Config.ProfilePath)
		if err != nil {
			return err
		}
		items := make([]tui.Item, 0, len(p.Registries))
		for _, r := range p.Registries {
			items = append(items, registryMgrItem{name: r.Name, url: r.URL})
		}

		cfg := tui.Config{
			Theme:      app.UI.Theme(),
			Version:    resolveVersion().Version,
			Section:    app.headerSection("Registries"),
			Items:      items,
			LeftTitle:  "REGISTRIES",
			RightTitle: "DETAIL",
			EmptyMsg:   "no registries — press a to add one",
			Actions: []tui.ActionSpec{
				{Key: "l", Label: "login", Action: "login"},
				{Key: "a", Label: "add", Action: "add"},
				{Key: "e", Label: "rename", Action: "rename"},
				{Key: "o", Label: "logout", Action: "logout"},
				{Key: "x", Label: "remove", Action: "remove"},
				{Key: "r", Label: "refresh", Action: "refresh"},
			},
		}
		if fromDashboard {
			cfg.BackKey = "esc"
			cfg.BackLabel = "back"
		}

		res, err := tui.RunListDetail(cfg)
		if err != nil {
			return err
		}
		if res.Quit {
			return nil
		}

		sel, hasSel := res.Item.(registryMgrItem)
		switch res.Action {
		case "add":
			if err := runRegistryAddInteractive(app); err != nil {
				return err
			}
		case "rename":
			if !hasSel {
				continue
			}
			newName, err := app.Prompt.Text("New name for "+sel.name, "")
			if err != nil {
				return err
			}
			if strings.TrimSpace(newName) == "" {
				continue
			}
			if err := runRegistryRename(app, sel.name, newName); err != nil {
				return err
			}
		case "login":
			if !hasSel {
				continue
			}
			tok, err := app.Prompt.Secret("Auth token for " + sel.name)
			if err != nil {
				return err
			}
			if strings.TrimSpace(tok) == "" {
				continue
			}
			if _, err := secrets.SetRegistryTokenFor(sel.name, tok); err != nil {
				return err
			}
		case "logout":
			if !hasSel {
				continue
			}
			if err := secrets.DeleteRegistryTokenFor(sel.name); err != nil {
				return err
			}
		case "remove":
			if !hasSel {
				continue
			}
			ok, err := app.Prompt.Confirm("Remove registry "+sel.name+"?", false)
			if err != nil {
				return err
			}
			if ok {
				if err := runRegistryRemove(app, sel.name); err != nil {
					return err
				}
			}
		case "refresh":
			if err := refreshAllRegistries(app); err != nil {
				return err
			}
		}
	}
}

// refreshAllRegistries forces a network refresh of every configured registry's
// cache, ignoring per-registry failures (they'll surface on next load).
func refreshAllRegistries(app *App) error {
	for _, r := range app.resolvedRegistries() {
		_, _, _ = app.fetcherForRegistry(r).Refresh()
	}
	return nil
}
