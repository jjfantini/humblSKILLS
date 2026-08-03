package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
)

// errCancelled means the user declined a destructive action. Callers report it
// as a normal outcome, not a failure.
var errCancelled = errors.New("cancelled")

// confirmDestructive gates anything that deletes user-owned files: uninstall,
// a --force reinstall (which bypasses the preserve list), and skillset prune.
//
// Three rules, applied uniformly so the CLI and the TUI can't drift apart:
//
//   - --yes is consent. Proceed.
//   - On a TTY, show exactly what will be destroyed and require an explicit
//     yes. The default is No.
//   - Non-interactive without --yes is an error, never an implied yes. A pipe
//     or --json run has nobody to answer the question, and these paths delete
//     files (LLM-maintained memory, raw notes) that no registry can restore.
//     detail names them in the error, since the styled panel can't be printed
//     there without corrupting --json output.
func confirmDestructive(app *App, title, prompt, detail string, lines []string) error {
	if app.Config.Yes {
		return nil
	}
	if !app.Prompt.Interactive || app.Config.JSON {
		msg := title
		if detail != "" {
			msg += ": " + detail
		}
		return fmt.Errorf("%s — re-run with --yes to confirm", msg)
	}
	ok, err := tui.ConfirmWithSummary(app.UI.Theme(), title, prompt, lines, false, true)
	if err != nil {
		return err
	}
	if !ok {
		return errCancelled
	}
	return nil
}

// confirmForce gates a --force run. --force is the documented way to discard
// local customizations, but it used to do so with no prompt at all, on skills
// whose whole point is accumulating user-owned memory.
//
// Nothing at risk (skills not installed yet, or no preserved files on disk)
// means no prompt: --force is then just a plain reinstall.
func confirmForce(app *App, action string, skills []string) error {
	m, err := manifest.Load(app.Config.ManifestPath)
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}
	var entries []*manifest.Installation
	for _, s := range skills {
		entries = append(entries, m.FindAll(s)...)
	}
	atRisk := preservedAcrossStores(entries)
	if len(atRisk) == 0 {
		return nil
	}

	theme := app.UI.Theme()
	names := make([]string, 0, len(atRisk))
	for skill := range atRisk {
		names = append(names, skill)
	}
	sort.Strings(names)

	lines := make([]string, 0, len(atRisk)*3)
	for _, skill := range names {
		lines = append(lines, theme.Name.Render(skill))
		for _, p := range atRisk[skill] {
			lines = append(lines, "  "+theme.Warn.Render("overwrites")+"  "+theme.Detail.Render(p))
		}
	}
	lines = append(lines, "",
		theme.Detail.Render("These are user-owned files (metadata.preserve). Without --force they survive."))

	return confirmDestructive(app,
		action,
		"Overwrite them from the registry?",
		flattenPreserved(atRisk, "overwrites"),
		lines,
	)
}

// preservedAcrossStores collects the user-owned preserved paths held by the
// canonical stores behind these installations, deduplicated and labelled by
// skill. It is what turns "are you sure?" into a list of files.
func preservedAcrossStores(entries []*manifest.Installation) map[string][]string {
	out := map[string][]string{}
	seen := map[string]struct{}{}
	for _, e := range entries {
		if e.StorePath == "" {
			continue
		}
		if _, done := seen[e.StorePath]; done {
			continue
		}
		seen[e.StorePath] = struct{}{}
		if at := install.PreservedAtRisk(e.StorePath, nil); len(at) > 0 {
			out[e.Skill] = append(out[e.Skill], at...)
		}
	}
	return out
}

// flattenPreserved renders the map from preservedAcrossStores as one plain
// line for a non-interactive error message. verb is the caller's word for what
// happens to the files ("deletes", "overwrites") — the two destructive paths do
// different things and the message shouldn't blur them. Returns "" when nothing
// is at risk, which lets callers skip the prompt entirely.
func flattenPreserved(bySkill map[string][]string, verb string) string {
	if len(bySkill) == 0 {
		return ""
	}
	parts := make([]string, 0, len(bySkill))
	for skill, paths := range bySkill {
		parts = append(parts, skill+" ("+strings.Join(paths, ", ")+")")
	}
	sort.Strings(parts)
	return verb + " user-owned files: " + strings.Join(parts, "; ")
}
