package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/textutil"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
)

// doctor is split across three files:
//   - doctor.go        (this file): command wiring + run orchestration
//   - doctor_report.go: the report data model and the logic that gathers it
//   - doctor_view.go:   TUI list items + non-TTY static rendering

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
	useTUI := tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes)

	var report doctorReport
	var err error
	if useTUI {
		// Build the report inside its own alt-screen loading spinner, not on
		// the exposed terminal buffer — see loadingModel's doc comment for
		// why that matters for avoiding the alt-screen "flash".
		report, err = tui.RunWithLoading(app.UI.Theme(), "scanning environment…", func() (doctorReport, error) {
			return buildDoctorReport(app)
		})
	} else {
		err = tui.RunWithSpinner(app.UI.Theme(), "scanning environment…", func() error {
			report, err = buildDoctorReport(app)
			return err
		})
	}
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

	if useTUI {
		for {
			rerun, err := runDoctorTUI(app, report)
			if err != nil {
				return err
			}
			if !rerun {
				break
			}
			report, err = tui.RunWithLoading(app.UI.Theme(), "rescanning…", func() (doctorReport, error) {
				return buildDoctorReport(app)
			})
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
	items = append(items, manifestItem{m: r.Manifest, issue: firstIssue(r.Issues, "manifest:")})
	for _, rr := range r.Registries {
		items = append(items, registryItem{reg: rr})
	}
	items = append(items,
		updatesItem{u: r.Updates},
		evalItem{e: r.Eval},
	)

	detected := 0
	for _, a := range r.Adapters {
		if a.Detected {
			detected++
		}
	}
	totalSkills, regErrs := 0, 0
	for _, rr := range r.Registries {
		totalSkills += rr.Skills
		if rr.Error != "" {
			regErrs++
		}
	}
	localMeta := func(_ []tui.Item, _ int) string {
		parts := []string{
			fmt.Sprintf("%d / %d detected", detected, len(r.Adapters)),
		}
		parts = append(parts, fmt.Sprintf("%d skills", totalSkills))
		if regErrs > 0 {
			parts = append(parts, fmt.Sprintf("%d registry error%s", regErrs, textutil.Plural(regErrs)))
		}
		if r.Updates.Count > 0 {
			parts = append(parts, fmt.Sprintf("%d update%s", r.Updates.Count, textutil.Plural(r.Updates.Count)))
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
