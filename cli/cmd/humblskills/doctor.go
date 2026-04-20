package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/evalruntime"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/workspace"
	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/secrets"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

type doctorReport struct {
	Adapters []adapterReport `json:"adapters"`
	Manifest manifestReport  `json:"manifest"`
	Registry registryReport  `json:"registry"`
	Updates  updatesReport   `json:"updates"`
	Eval     evalReport      `json:"eval"`
	Issues   []string        `json:"issues,omitempty"`
}

// evalReport is the eval-prerequisite block: per-runner availability,
// workspace writability, per-provider API key source (never the value),
// and the count of installed skills with evals/ on disk.
type evalReport struct {
	Runners       []runnerCheck  `json:"runners"`
	Workspace     workspaceCheck `json:"workspace"`
	APIKeys       []apiKeyCheck  `json:"api_keys"`
	DefaultRunner string         `json:"default_runner"`
	EvalSkills    int            `json:"eval_skills"`
}

type runnerCheck struct {
	Name        string `json:"name"`
	Available   bool   `json:"available"`
	Version     string `json:"version,omitempty"`
	Reason      string `json:"reason,omitempty"`
	Fix         string `json:"fix,omitempty"`
	RequiresKey string `json:"requires_key,omitempty"`
}

type workspaceCheck struct {
	Path           string `json:"path"`
	Exists         bool   `json:"exists"`
	Writable       bool   `json:"writable"`
	SizeBytes      int64  `json:"size_bytes"`
	IterationCount int    `json:"iteration_count"`
	SkillsWithRuns int    `json:"skills_with_runs"`
}

type apiKeyCheck struct {
	Provider string `json:"provider"`
	Present  bool   `json:"present"`
	Source   string `json:"source"` // "env" | "keyring" | "file" | "absent"
}

type updatesReport struct {
	Count  int      `json:"count"`
	Skills []string `json:"skills,omitempty"`
}

type adapterReport struct {
	Name     string         `json:"name"`
	Detected bool           `json:"detected"`
	Reason   string         `json:"reason"`
	Targets  []targetReport `json:"targets"`
}

type targetReport struct {
	Scope    string `json:"scope"`
	Path     string `json:"path"`
	Writable bool   `json:"writable"`
}

type manifestReport struct {
	Path          string `json:"path"`
	SchemaVersion int    `json:"schema_version"`
	Installs      int    `json:"installs"`
}

type registryReport struct {
	URL       string        `json:"url"`
	Source    string        `json:"source"`
	Cached    bool          `json:"cached"`
	FetchedAt time.Time     `json:"fetched_at,omitempty"`
	Age       time.Duration `json:"age_seconds,omitempty"`
	Skills    int           `json:"skills"`
	DepIssues []string      `json:"dep_issues,omitempty"`
	Error     string        `json:"error,omitempty"`
}

func newDoctorCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check environment, targets, manifest, and registry health",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDoctor(app)
		},
	}
}

func runDoctor(app *App) error {
	report, err := buildDoctorReport(app)
	if err != nil {
		return err
	}

	if app.Config.JSON {
		if err := app.UI.JSON(report); err != nil {
			return err
		}
		if hasFailures(report) {
			return errDoctorFailed
		}
		return nil
	}

	if tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes) {
		for {
			rerun, err := runDoctorTUI(app, report)
			if err != nil {
				return err
			}
			if !rerun {
				break
			}
			report, err = buildDoctorReport(app)
			if err != nil {
				return err
			}
		}
	} else {
		printDoctorStatic(app, report)
	}

	if hasFailures(report) {
		return errDoctorFailed
	}
	return nil
}

func buildDoctorReport(app *App) (doctorReport, error) {
	report := doctorReport{}
	adapterList, err := app.Adapters()
	if err != nil {
		return report, fmt.Errorf("load adapters: %w", err)
	}

	results := adapters.Detect(adapterList)
	byName := make(map[string]adapters.Adapter, len(adapterList))
	for _, a := range adapterList {
		byName[a.Name] = a
	}
	for _, r := range results {
		ar := adapterReport{Name: r.Name, Detected: r.Detected, Reason: r.Reason}
		for _, t := range byName[r.Name].Targets() {
			ar.Targets = append(ar.Targets, targetReport{Scope: t.Scope, Path: t.Path, Writable: t.Writable})
		}
		report.Adapters = append(report.Adapters, ar)
	}

	mpath := app.Config.ManifestPath
	m, mErr := manifest.Load(mpath)
	if mErr != nil {
		report.Issues = append(report.Issues, fmt.Sprintf("manifest: %s", mErr))
		report.Manifest = manifestReport{Path: mpath}
	} else {
		report.Manifest = manifestReport{
			Path:          mpath,
			SchemaVersion: m.SchemaVersion,
			Installs:      len(m.Installations),
		}
	}

	f := registry.NewFetcher(app.Config.RegistryURL, app.Config.CacheDir)
	var (
		reg    *registry.Registry
		origin registry.Origin
		rErr   error
	)
	// Don't spin during the TUI; buildDoctorReport is called outside alt-screen.
	_ = tui.RunWithSpinner(app.UI.Theme(), "loading registry…", func() error {
		reg, origin, rErr = f.Load()
		return nil
	})
	rr := registryReport{URL: app.Config.RegistryURL, Source: string(origin)}
	if rErr != nil {
		rr.Error = rErr.Error()
		report.Issues = append(report.Issues, fmt.Sprintf("registry: %s", rErr))
	} else {
		rr.Skills = len(reg.Skills)
		for _, issue := range registry.ValidateDeps(reg) {
			rr.DepIssues = append(rr.DepIssues, issue.Error())
		}
	}
	info := f.Inspect()
	rr.Cached = info.Exists
	rr.FetchedAt = info.FetchedAt
	rr.Age = info.Age
	report.Registry = rr

	if mErr == nil && rErr == nil {
		plans := install.PlanUpdates(reg, m, nil)
		report.Updates.Count = len(plans)
		for _, p := range plans {
			report.Updates.Skills = append(report.Updates.Skills, p.Skill)
		}
	}

	report.Eval = buildEvalReport(app)
	return report, nil
}

// buildEvalReport collects per-runner availability + API-key presence +
// workspace writability for the doctor report. Never touches the wire -
// runner DoctorCheck is designed to be a fast local probe.
func buildEvalReport(app *App) evalReport {
	store, _ := secrets.NewStore("")
	reg := evalruntime.DefaultRegistry(store)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	det := reg.Detect(ctx)

	er := evalReport{}
	for _, d := range det {
		er.Runners = append(er.Runners, runnerCheck{
			Name:        d.Name,
			Available:   d.Check.Available,
			Version:     d.Check.Version,
			Reason:      d.Check.Reason,
			Fix:         d.Check.Fix,
			RequiresKey: d.Check.RequiresKey,
		})
		if er.DefaultRunner == "" && d.Check.Available {
			er.DefaultRunner = d.Name
		}
	}
	// API keys.
	if store != nil {
		for _, p := range secrets.Providers() {
			_, src, err := store.Get(p.Name)
			present := err == nil && src != secrets.SourceAbsent
			er.APIKeys = append(er.APIKeys, apiKeyCheck{
				Provider: p.Name,
				Present:  present,
				Source:   string(src),
			})
		}
	}
	// Workspace.
	wsRoot := resolveWorkspace(app, "")
	ws := workspaceCheck{Path: wsRoot}
	if fi, err := os.Stat(wsRoot); err == nil && fi.IsDir() {
		ws.Exists = true
		probe := filepath.Join(wsRoot, ".writable-probe")
		if err := os.WriteFile(probe, []byte("ok"), 0o644); err == nil {
			_ = os.Remove(probe)
			ws.Writable = true
		}
		ws.SizeBytes, _ = workspace.SizeBytes(wsRoot)
		skills, _ := workspace.ListSkills(wsRoot)
		ws.SkillsWithRuns = len(skills)
		for _, s := range skills {
			if n, _ := workspace.MaxIterationN(wsRoot, s); n > 0 {
				ws.IterationCount += n
			}
		}
	} else {
		// Non-existent is expected on first run; writability reflects the
		// parent dir's state so we can surface real permission issues.
		if parent := filepath.Dir(wsRoot); parent != "" {
			if _, err := os.Stat(parent); err == nil {
				ws.Writable = true
			}
		}
	}
	er.Workspace = ws
	// Count installed skills with evals/ directories.
	m, err := manifest.Load(app.Config.ManifestPath)
	if err == nil && m != nil {
		for _, inst := range m.Installations {
			if _, err := os.Stat(filepath.Join(inst.Path, "evals")); err == nil {
				er.EvalSkills++
			}
		}
	}
	return er
}

var errDoctorFailed = errors.New("doctor found issues")

func hasFailures(r doctorReport) bool {
	if len(r.Issues) > 0 {
		return true
	}
	if r.Registry.Error != "" || len(r.Registry.DepIssues) > 0 {
		return true
	}
	// Eval: default runner must be available for CI to claim "healthy".
	if r.Eval.DefaultRunner == "" && len(r.Eval.Runners) > 0 {
		return true
	}
	return false
}

// runDoctorTUI drives the two-pane listdetail surface for `doctor`. Returns
// rerun=true if the user pressed `r` (rescan) so the caller can rebuild the
// report and re-enter the TUI.
func runDoctorTUI(app *App, r doctorReport) (bool, error) {
	th := app.UI.Theme()
	ver := resolveVersion().Version

	items := make([]tui.Item, 0, len(r.Adapters)+4)
	for _, a := range r.Adapters {
		items = append(items, adapterItem{a: a})
	}
	items = append(items,
		manifestItem{m: r.Manifest, issue: firstIssue(r.Issues, "manifest:")},
		registryItem{reg: r.Registry},
		updatesItem{u: r.Updates},
		evalItem{e: r.Eval},
	)

	detected := 0
	for _, a := range r.Adapters {
		if a.Detected {
			detected++
		}
	}
	localMeta := func(_ []tui.Item, _ int) string {
		parts := []string{
			fmt.Sprintf("%d / %d detected", detected, len(r.Adapters)),
		}
		if r.Registry.Error == "" {
			parts = append(parts, fmt.Sprintf("%d skills", r.Registry.Skills))
		} else {
			parts = append(parts, "registry error")
		}
		if r.Updates.Count > 0 {
			parts = append(parts, fmt.Sprintf("%d update%s", r.Updates.Count, plural(r.Updates.Count)))
		}
		return strings.Join(parts, " · ")
	}
	meta := localMeta
	if app.Nav.Crumb != "" {
		status := app.Nav.Status
		meta = func(_ []tui.Item, _ int) string {
			return tui.RenderStatusMeta(th, status)
		}
	}

	res, err := tui.RunListDetail(tui.Config{
		Theme:      th,
		Version:    ver,
		Section:    app.headerSection("Doctor"),
		Meta:       meta,
		Items:      items,
		LeftTitle:  "CHECKS",
		RightTitle: "DETAIL",
		Actions: []tui.ActionSpec{
			{Key: "r", Label: "rescan", Action: "rescan"},
		},
	})
	if err != nil {
		return false, err
	}
	return res.Action == "rescan", nil
}

func firstIssue(issues []string, prefix string) string {
	for _, i := range issues {
		if strings.HasPrefix(i, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(i, prefix))
		}
	}
	return ""
}

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
				i.e.Workspace.IterationCount, plural(i.e.Workspace.IterationCount),
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
			dot, th.Name.Render(padRight(r.Name, 14)), status, th.Detail.Render(firstNonEmptyStr(r.Version, r.Reason))))
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
			th.Detail.Render(fmt.Sprintf("(%d install%s)", r.Manifest.Installs, plural(r.Manifest.Installs))),
		)
	} else {
		app.UI.Detail("  %s (not yet created)", r.Manifest.Path)
	}

	app.UI.Section("Registry")
	app.UI.Info("  %s %s", th.Label.Render("url"), th.Name.Render(r.Registry.URL))
	if r.Registry.Error != "" {
		app.UI.Error("registry unreachable: %s", r.Registry.Error)
	} else {
		app.UI.Success("%d skill%s available (via %s)", r.Registry.Skills, plural(r.Registry.Skills), r.Registry.Source)
		if r.Registry.Cached {
			app.UI.Detail("  cache fetched %s ago", r.Registry.Age.Round(time.Second))
		}
		for _, issue := range r.Registry.DepIssues {
			app.UI.Warn("dep issue: %s", issue)
		}
	}

	if r.Updates.Count > 0 {
		app.UI.Section("Updates")
		app.UI.Warn("%d skill%s can be updated — run 'humblskills update'", r.Updates.Count, plural(r.Updates.Count))
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

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
