package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/mirrors"
)

// defaultSkillsDir resolves the skills tree to inspect. Mirrors are checked
// against a source checkout, not installed copies: an installed skill has no
// upstream to re-distil into.
func defaultSkillsDir(flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := wd; ; {
		candidate := filepath.Join(dir, "skills")
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no skills/ directory found from %s (pass --skills-dir)", wd)
}

func newMirrorsCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mirrors",
		Short: "Check mirrored skills against their upstream",
		Long: "Detect drift between skills that declare an `upstream:` block and the " +
			"upstream they mirror.\n\n" +
			"Detection is deterministic; re-distillation is not, and is never automated. " +
			"`mirrors plan` emits a work order for a human or agent to execute.",
	}
	cmd.AddCommand(newMirrorsCheckCmd(app), newMirrorsPlanCmd(app))
	return cmd
}

type mirrorsFlags struct {
	skillsDir string
	outDir    string
}

func newMirrorsCheckCmd(app *App) *cobra.Command {
	var f mirrorsFlags
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Report which mirrored skills have drifted from upstream",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := defaultSkillsDir(f.skillsDir)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			results, err := mirrors.Check(ctx, dir, nil)
			if err != nil {
				return err
			}
			_ = mirrors.Stamp(mirrorsStampPath(app))

			if app.Config.JSON {
				return app.UI.JSON(results)
			}
			if len(results) == 0 {
				app.UI.Info("no mirrored skills found (none declare an `upstream:` block)")
				return nil
			}
			for _, r := range results {
				line := fmt.Sprintf("%-28s %-10s %s", r.Skill, r.Status, r.Reason)
				switch r.Status {
				case mirrors.StatusRewritten, mirrors.StatusDrifted:
					app.UI.Warn("%s", line)
				case mirrors.StatusUnknown:
					app.UI.Info("%s", line)
				default:
					app.UI.Info("%s", line)
				}
			}
			if s := mirrors.Summary(results); s != "" {
				app.UI.Info("")
				app.UI.Warn("%s", s)
				app.UI.Info("next: humblskills mirrors plan <skill>")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&f.skillsDir, "skills-dir", "", "path to the skills/ directory (default: nearest ancestor)")
	return cmd
}

func newMirrorsPlanCmd(app *App) *cobra.Command {
	var f mirrorsFlags
	cmd := &cobra.Command{
		Use:   "plan <skill>",
		Short: "Write a re-sync work order for a drifted mirror",
		Long: "Emit everything a re-sync needs and that is otherwise reconstructed by hand: " +
			"the two files to diff, the wiki concepts that cite the preserved copy, the " +
			"declared deltas that must survive, and the completion checklist.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := defaultSkillsDir(f.skillsDir)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			results, err := mirrors.Check(ctx, dir, nil)
			if err != nil {
				return err
			}
			var target *mirrors.Result
			for i := range results {
				if results[i].Skill == args[0] {
					target = &results[i]
					break
				}
			}
			if target == nil {
				return fmt.Errorf("%s is not a mirrored skill (no `upstream:` block)", args[0])
			}
			if target.Status == mirrors.StatusCurrent {
				app.UI.Success("%s is already current with upstream — nothing to plan", args[0])
				return nil
			}
			if target.Status == mirrors.StatusUnknown {
				return fmt.Errorf("%s: %s", args[0], target.Reason)
			}

			out := f.outDir
			if out == "" {
				out = filepath.Join(filepath.Dir(dir), ".mirror-sync")
			}
			a, err := mirrors.WritePlan(*target, out)
			if err != nil {
				return err
			}
			if app.Config.JSON {
				return app.UI.JSON(a)
			}
			app.UI.Success("wrote re-sync plan for %s", args[0])
			app.UI.Info("  plan:     %s", a.PlanPath)
			app.UI.Info("  incoming: %s", a.IncomingPath)
			app.UI.Info("  affected: %d concept(s)", len(a.Affected))
			app.UI.Info("")
			app.UI.Info("hand the plan to an agent, or work it yourself — it is a prompt, not a patch")
			return nil
		},
	}
	cmd.Flags().StringVar(&f.skillsDir, "skills-dir", "", "path to the skills/ directory (default: nearest ancestor)")
	cmd.Flags().StringVar(&f.outDir, "out-dir", "", "where to write the plan (default: <repo>/.mirror-sync)")
	return cmd
}

func mirrorsStampPath(app *App) string {
	return filepath.Join(app.Config.CacheDir, "mirrors-checked")
}
