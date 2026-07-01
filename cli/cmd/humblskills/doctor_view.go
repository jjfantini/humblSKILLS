package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/workspace"
	"github.com/jjfantini/humblSKILLS/cli/internal/textutil"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// This file is the doctor presentation layer: the listdetail Item adapters
// used by the TUI plus the non-TTY static renderer. The report data model and
// its builders live in doctor_report.go.

// ---- items -----------------------------------------------------------------

type adapterItem struct{ a adapterReport }

func (a adapterItem) Key() string         { return a.a.Name }
func (a adapterItem) FilterValue() string { return a.a.Name + " " + a.a.Reason }
func (a adapterItem) NaturalWidth(th *ui.Theme) int {
	badge := tui.Badge(th, tui.BadgeMissing, "detected") // both "detected" and "missing" are 8 glyphs; pad puts them at identical width
	return rowNaturalWidth(a.a.Name, lipgloss.Width(badge))
}
func (a adapterItem) Row(th *ui.Theme, width int, selected bool) string {
	var dot, badge string
	if a.a.Detected {
		dot = th.DotOK.Render("●")
		badge = tui.Badge(th, tui.BadgeDetected, "detected")
	} else {
		dot = th.DotNo.Render("●")
		badge = tui.Badge(th, tui.BadgeMissing, "missing")
	}
	name := rowName(th, a.a.Name, selected, a.a.Detected)
	return rowWithTrailingBadge(dot+" "+name, badge, width)
}
func (a adapterItem) Detail(th *ui.Theme, width int) string {
	var sb strings.Builder
	title := th.DetailTitle.Render(a.a.Name)
	sub := th.DetailSub.Render(ternary(a.a.Detected, "detected", "not detected"))
	sb.WriteString(title + "  " + sub + "\n\n")

	sb.WriteString(kvRow(th, "status", ternary(a.a.Detected,
		th.Success.Render("detected"),
		th.Error.Render("missing"))))
	if a.a.Reason != "" {
		sb.WriteString(kvRow(th, "matched on", th.KVValue.Render(a.a.Reason)))
	}

	if len(a.a.Targets) > 0 {
		sb.WriteString("\n")
		sb.WriteString(th.SectionTitle.Render("PATHS"))
		sb.WriteString("\n")
		for i, t := range a.a.Targets {
			if i > 0 {
				sb.WriteString(tui.DashedRule(th, width) + "\n")
			}
			sb.WriteString(pathRow(th, t) + "\n")
		}
	}
	return sb.String()
}

type manifestItem struct {
	m     manifestReport
	issue string
}

func (m manifestItem) Key() string         { return "manifest" }
func (m manifestItem) FilterValue() string { return "manifest " + m.m.Path }
func (m manifestItem) NaturalWidth(th *ui.Theme) int {
	badge := tui.Badge(th, tui.BadgeMissing, "error") // "error" is the wider of {"ok","error"}
	return rowNaturalWidth("manifest", lipgloss.Width(badge))
}
func (m manifestItem) Row(th *ui.Theme, width int, selected bool) string {
	ok := m.issue == ""
	var dot, badge string
	if ok {
		dot = th.DotOK.Render("●")
		badge = tui.Badge(th, tui.BadgeDetected, "ok")
	} else {
		dot = th.DotNo.Render("●")
		badge = tui.Badge(th, tui.BadgeMissing, "error")
	}
	return rowWithTrailingBadge(dot+" "+rowName(th, "manifest", selected, true), badge, width)
}
func (m manifestItem) Detail(th *ui.Theme, width int) string {
	var sb strings.Builder
	sb.WriteString(th.DetailTitle.Render("manifest") + "  " +
		th.DetailSub.Render(m.m.Path) + "\n\n")
	sb.WriteString(kvRow(th, "schema", th.KVValue.Render(fmt.Sprintf("v%d", m.m.SchemaVersion))))
	sb.WriteString(kvRow(th, "installs", th.KVValue.Render(fmt.Sprintf("%d", m.m.Installs))))
	sb.WriteString(kvRow(th, "path", th.KVValue.Render(m.m.Path)))
	if m.issue != "" {
		sb.WriteString("\n" + th.Error.Render("! "+m.issue) + "\n")
	} else if _, err := os.Stat(m.m.Path); err != nil {
		sb.WriteString("\n" + th.Detail.Render("(file not yet created)") + "\n")
	}
	_ = width
	return sb.String()
}

type registryItem struct{ reg registryReport }

func (r registryItem) Key() string         { return "registry" }
func (r registryItem) FilterValue() string { return "registry " + r.reg.URL }
func (r registryItem) NaturalWidth(th *ui.Theme) int {
	var badge string
	switch {
	case r.reg.Error != "":
		badge = tui.Badge(th, tui.BadgeMissing, "unreachable")
	case len(r.reg.DepIssues) > 0:
		badge = tui.Badge(th, tui.BadgeRO, "issues")
	default:
		badge = tui.Badge(th, tui.BadgeDetected, fmt.Sprintf("%d skills", r.reg.Skills))
	}
	return rowNaturalWidth("registry", lipgloss.Width(badge))
}
func (r registryItem) Row(th *ui.Theme, width int, selected bool) string {
	ok := r.reg.Error == "" && len(r.reg.DepIssues) == 0
	var dot, badge string
	if r.reg.Error != "" {
		dot = th.DotNo.Render("●")
		badge = tui.Badge(th, tui.BadgeMissing, "unreachable")
	} else if len(r.reg.DepIssues) > 0 {
		dot = th.DotWarn.Render("●")
		badge = tui.Badge(th, tui.BadgeRO, "issues")
	} else {
		dot = th.DotOK.Render("●")
		badge = tui.Badge(th, tui.BadgeDetected, fmt.Sprintf("%d skills", r.reg.Skills))
	}
	return rowWithTrailingBadge(dot+" "+rowName(th, "registry", selected, ok), badge, width)
}
func (r registryItem) Detail(th *ui.Theme, width int) string {
	var sb strings.Builder
	sb.WriteString(th.DetailTitle.Render("registry") + "  " +
		th.DetailSub.Render(r.reg.URL) + "\n\n")
	sb.WriteString(kvRow(th, "source", th.KVValue.Render(r.reg.Source)))
	sb.WriteString(kvRow(th, "skills", th.KVValue.Render(fmt.Sprintf("%d", r.reg.Skills))))
	if r.reg.Cached {
		sb.WriteString(kvRow(th, "cached", th.KVValue.Render(r.reg.Age.Round(time.Second).String()+" ago")))
	}
	if r.reg.Error != "" {
		sb.WriteString("\n" + th.Error.Render("! "+r.reg.Error) + "\n")
	}
	for _, issue := range r.reg.DepIssues {
		sb.WriteString(th.Warn.Render("! "+issue) + "\n")
	}
	_ = width
	return sb.String()
}

type updatesItem struct{ u updatesReport }

func (u updatesItem) Key() string         { return "updates" }
func (u updatesItem) FilterValue() string { return "updates" }
func (u updatesItem) NaturalWidth(th *ui.Theme) int {
	var badge string
	if u.u.Count == 0 {
		badge = tui.Badge(th, tui.BadgeDetected, "up-to-date")
	} else {
		badge = tui.Badge(th, tui.BadgeRO, fmt.Sprintf("%d outdated", u.u.Count))
	}
	return rowNaturalWidth("updates", lipgloss.Width(badge))
}
func (u updatesItem) Row(th *ui.Theme, width int, selected bool) string {
	var dot, badge string
	if u.u.Count == 0 {
		dot = th.DotOK.Render("●")
		badge = tui.Badge(th, tui.BadgeDetected, "up-to-date")
	} else {
		dot = th.DotWarn.Render("●")
		badge = tui.Badge(th, tui.BadgeRO, fmt.Sprintf("%d outdated", u.u.Count))
	}
	return rowWithTrailingBadge(dot+" "+rowName(th, "updates", selected, true), badge, width)
}
func (u updatesItem) Detail(th *ui.Theme, width int) string {
	var sb strings.Builder
	sb.WriteString(th.DetailTitle.Render("updates") + "\n\n")
	if u.u.Count == 0 {
		sb.WriteString(th.Detail.Render("every installed skill matches the registry version."))
		return sb.String()
	}
	sb.WriteString(kvRow(th, "drifted", th.KVValue.Render(fmt.Sprintf("%d", u.u.Count))))
	sb.WriteString("\n")
	for _, name := range u.u.Skills {
		sb.WriteString("  " + th.KVValue.Render("• "+name) + "\n")
	}
	sb.WriteString("\n" + th.Detail.Render("run 'humblskills update' to apply"))
	_ = width
	return sb.String()
}

// evalItem renders the Eval block in the doctor TUI. Mirrors updatesItem's
// shape so layout stays uniform across the listdetail pane.
type evalItem struct{ e evalReport }

func (i evalItem) Key() string         { return "eval" }
func (i evalItem) FilterValue() string { return "eval" }
func (i evalItem) NaturalWidth(th *ui.Theme) int {
	badge := tui.Badge(th, tui.BadgeMissing, "no runner")
	return rowNaturalWidth("eval", lipgloss.Width(badge))
}
func (i evalItem) Row(th *ui.Theme, width int, selected bool) string {
	var dot, badge string
	if i.e.DefaultRunner != "" {
		dot = th.DotOK.Render("●")
		badge = tui.Badge(th, tui.BadgeDetected, i.e.DefaultRunner)
	} else {
		dot = th.DotNo.Render("●")
		badge = tui.Badge(th, tui.BadgeMissing, "no runner")
	}
	return rowWithTrailingBadge(dot+" "+rowName(th, "eval", selected, i.e.DefaultRunner != ""), badge, width)
}
func (i evalItem) Detail(th *ui.Theme, width int) string {
	var sb strings.Builder
	sb.WriteString(th.DetailTitle.Render("eval") + "\n\n")
	sb.WriteString(kvRow(th, "workspace", th.KVValue.Render(i.e.Workspace.Path)))
	if i.e.Workspace.Exists {
		sb.WriteString(kvRow(th, "size",
			th.KVValue.Render(fmt.Sprintf("%d iteration%s · %s",
				i.e.Workspace.IterationCount, textutil.Plural(i.e.Workspace.IterationCount),
				workspace.HumanSize(i.e.Workspace.SizeBytes)))))
	}
	sb.WriteString(kvRow(th, "default", th.KVValue.Render(nonEmptyOrDash(i.e.DefaultRunner))))
	sb.WriteString("\n")
	sb.WriteString(th.SectionTitle.Render("RUNNERS") + "\n")
	for _, r := range i.e.Runners {
		dot := th.DotOK.Render("●")
		status := th.Success.Render("ready")
		if !r.Available {
			dot = th.DotNo.Render("●")
			status = th.Error.Render("missing")
		}
		sb.WriteString(fmt.Sprintf("  %s %s  %s  %s\n",
			dot, th.Name.Render(padRight(r.Name, 14)), status, th.Detail.Render(textutil.FirstNonBlank(r.Version, r.Reason))))
		if !r.Available && r.Fix != "" {
			sb.WriteString("    " + th.Detail.Render("fix: "+r.Fix) + "\n")
		}
	}
	if len(i.e.APIKeys) > 0 {
		sb.WriteString("\n" + th.SectionTitle.Render("API KEYS") + "\n")
		for _, k := range i.e.APIKeys {
			state := th.Error.Render("absent")
			if k.Present {
				state = th.Success.Render("present (" + k.Source + ")")
			}
			sb.WriteString(fmt.Sprintf("  %s %s\n", th.Name.Render(padRight(k.Provider, 14)), state))
		}
	}
	_ = width
	return sb.String()
}

func nonEmptyOrDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// ---- shared detail helpers -------------------------------------------------

// rowName renders an item's name in the correct palette for the left pane:
// selected → magenta bold; dimmed (missing/unavailable) → comment; otherwise
// fg.
func rowName(th *ui.Theme, name string, selected, enabled bool) string {
	switch {
	case selected:
		return th.RowSelected.Render(name)
	case !enabled:
		return th.RowDim.Render(name)
	default:
		return th.RowUnselected.Render(name)
	}
}

// rowNaturalWidth returns the natural display width of a left-pane row whose
// body reads `● <label>  <badge>`: 1 dot + 1 space + the label's display
// width + 2 gap + the badge's display width. Items pass this up through
// SizedItem.NaturalWidth so the two-pane model can size the left column to
// the actual content instead of clamping to a hard-coded width.
func rowNaturalWidth(label string, badgeWidth int) int {
	// 1 (dot) + 1 (space) + label + 2 (gap) + badge.
	return 1 + 1 + lipgloss.Width(label) + 2 + badgeWidth
}

// rowWithTrailingBadge lays out a left-anchored label and a right-anchored
// badge within `width` cells. When the badge can't fit it's dropped, but the
// label is still padded so every row ends at the same column — otherwise the
// divider snaps to the widest row and loses alignment.
func rowWithTrailingBadge(label, badge string, width int) string {
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

// kvRow formats one key/value pair as `  label    value` with the label padded
// to 11 cells (matching the design's 110px label column, in mono-font terms).
func kvRow(th *ui.Theme, key, value string) string {
	label := th.KVKey.Render(key)
	pad := 11 - lipgloss.Width(label)
	if pad < 1 {
		pad = 1
	}
	return "  " + label + strings.Repeat(" ", pad) + value + "\n"
}

// pathRow formats one target as `scope  [rw]  path`, matching the PATHS stack
// in the design's detail pane.
func pathRow(th *ui.Theme, t targetReport) string {
	scope := th.PathLabel.Render(padRight(t.Scope, 7))
	var rw string
	if t.Writable {
		rw = tui.Badge(th, tui.BadgeRW, "rw")
	} else {
		rw = tui.Badge(th, tui.BadgeRO, "ro")
	}
	path := th.PathValue.Render(t.Path)
	return "  " + scope + "  " + rw + "  " + path
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

func ternary(c bool, a, b string) string {
	if c {
		return a
	}
	return b
}

// ---- non-TTY static render -------------------------------------------------

// printDoctorStatic writes the report as stacked sections. Fits 80 cols, works
// in pipes, keeps colour.
func printDoctorStatic(app *App, r doctorReport) {
	th := app.UI.Theme()
	app.UI.Header("doctor")

	app.UI.Section("Adapters")
	anyDetected := false
	for _, a := range r.Adapters {
		if a.Detected {
			anyDetected = true
		}
		dot := th.DotOK.Render("●")
		badge := tui.Badge(th, tui.BadgeDetected, "detected")
		if !a.Detected {
			dot = th.DotNo.Render("●")
			badge = tui.Badge(th, tui.BadgeMissing, "missing")
		}
		fmt.Fprintf(app.UI.Out(), "  %s %s  %s\n", dot, th.DetailTitle.Render(a.Name), badge)
		if a.Reason != "" {
			fmt.Fprintf(app.UI.Out(), "    %s %s\n", th.KVKey.Render("matched on"), th.KVValue.Render(a.Reason))
		}
		for _, t := range a.Targets {
			fmt.Fprintln(app.UI.Out(), "    "+pathRow(th, t))
		}
	}
	if !anyDetected {
		app.UI.Warn("no agent platform detected — run inside a project that uses Claude Code, Cursor, etc.")
	}

	app.UI.Section("Manifest")
	if _, err := os.Stat(r.Manifest.Path); err == nil {
		app.UI.Info("  %s %s",
			th.Name.Render(r.Manifest.Path),
			th.Detail.Render(fmt.Sprintf("(%d install%s)", r.Manifest.Installs, textutil.Plural(r.Manifest.Installs))),
		)
	} else {
		app.UI.Detail("  %s (not yet created)", r.Manifest.Path)
	}

	app.UI.Section("Registry")
	app.UI.Info("  %s %s", th.Label.Render("url"), th.Name.Render(r.Registry.URL))
	if r.Registry.Error != "" {
		app.UI.Error("registry unreachable: %s", r.Registry.Error)
	} else {
		app.UI.Success("%d skill%s available (via %s)", r.Registry.Skills, textutil.Plural(r.Registry.Skills), r.Registry.Source)
		if r.Registry.Cached {
			app.UI.Detail("  cache fetched %s ago", r.Registry.Age.Round(time.Second))
		}
		for _, issue := range r.Registry.DepIssues {
			app.UI.Warn("dep issue: %s", issue)
		}
	}

	if r.Updates.Count > 0 {
		app.UI.Section("Updates")
		app.UI.Warn("%d skill%s can be updated - run 'humblskills update'", r.Updates.Count, textutil.Plural(r.Updates.Count))
		for _, name := range r.Updates.Skills {
			app.UI.Detail("  • %s", name)
		}
	}

	printEvalStatic(app, r.Eval)

	for _, i := range r.Issues {
		app.UI.Warn(i)
	}
}

// printEvalStatic renders the Eval doctor section in non-TTY mode. Mirrors
// the Adapters / Manifest / Registry / Updates stack. No emoji per house
// style - the same dot + badge pattern every other block uses.
func printEvalStatic(app *App, e evalReport) {
	th := app.UI.Theme()
	app.UI.Section("Eval")
	// Workspace.
	app.UI.Info("  %s %s",
		th.Label.Render("workspace"),
		th.Name.Render(e.Workspace.Path))
	if e.Workspace.Exists {
		app.UI.Detail("    %d iterations · %s",
			e.Workspace.IterationCount, workspace.HumanSize(e.Workspace.SizeBytes))
	} else {
		app.UI.Detail("    (will be created on first eval)")
	}
	// Runners.
	for _, rn := range e.Runners {
		var dot, badge string
		if rn.Available {
			dot = th.DotOK.Render("●")
			badge = tui.Badge(th, tui.BadgeDetected, "ready")
		} else {
			dot = th.DotNo.Render("●")
			badge = tui.Badge(th, tui.BadgeMissing, "missing")
		}
		fmt.Fprintf(app.UI.Out(), "  %s %s  %s %s\n",
			dot, th.DetailTitle.Render(rn.Name), badge, th.Detail.Render(rn.Version))
		if !rn.Available && rn.Fix != "" {
			app.UI.Detail("    fix: %s", rn.Fix)
		}
	}
	// API keys.
	if len(e.APIKeys) > 0 {
		app.UI.Detail("  api keys:")
		for _, k := range e.APIKeys {
			state := "absent"
			if k.Present {
				state = "present via " + k.Source
			}
			app.UI.Detail("    %s  %s", k.Provider, state)
		}
	}
	if e.DefaultRunner == "" {
		app.UI.Warn("no eval runner available — run `humblskills eval set-key anthropic` or install one of: claude, cursor-agent, codex")
	} else {
		app.UI.Detail("  default runner: %s", e.DefaultRunner)
	}
}
