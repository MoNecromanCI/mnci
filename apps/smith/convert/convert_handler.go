package convert

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// ErrBlocked is returned when the plan has findings that stop the conversion.
var ErrBlocked = errors.New("the conversion is blocked; see the report")

// Command is `smith convert`.
func Command(version string) *cobra.Command {
	var dryRun, asJSON, strict, skipInstall, allowDirty bool
	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert an Nx workspace (or one made by the old mnci) to smith",
		Long: `convert turns the Nx workspace in the current directory into a smith workspace:
projects keep their directories, targets become plugin tasks, package.json
scripts or smith.toml tasks, release settings and existing tags carry over, Nx is
removed and the CI pipelines are regenerated.

Run it with --dry-run first: the report lists what maps, what changes behaviour
and what needs you. Nothing is dropped silently. The workspace's dependencies
must be installed (npm ci), because Nx itself describes the projects.`,
		Example: "  smith convert --dry-run\n  smith convert",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			if _, err := os.Stat(filepath.Join(root, workspace.FileName)); err == nil {
				return fmt.Errorf("%s already exists: this workspace is already a smith workspace", workspace.FileName)
			}
			git := gitclient.New(root)
			if !git.IsRepository() {
				return errors.New("convert runs in a git repository, so every change it makes can be reviewed and undone")
			}
			if !dryRun && !allowDirty {
				if status, err := git.Run("status", "--porcelain"); err == nil && status != "" {
					return errors.New("the working tree has changes: commit or stash them first, so the conversion is one reviewable change (or pass --allow-dirty)")
				}
			}
			files, err := git.Files()
			if err != nil {
				return err
			}
			w, err := nxworkspace.Read(root, files)
			if err != nil {
				return err
			}
			tags, _ := git.Tags("")
			plan := convertplan.MakePlan(w, convertplan.Input{
				DirName: filepath.Base(root),
				Files:   files,
				Tags:    tags,
				Read: func(rel string) string {
					raw, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
					return string(raw)
				},
			})
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				if err := enc.Encode(plan); err != nil {
					return err
				}
			} else {
				fmt.Fprint(out, Report(plan, dryRun))
			}
			if plan.Blocked(strict) {
				return ErrBlocked
			}
			if dryRun {
				if !asJSON {
					fmt.Fprintln(out, "\nNothing changed (dry run). Run `smith convert` to apply.")
				}
				return nil
			}
			log := out
			if asJSON {
				log = cmd.ErrOrStderr()
			}
			fmt.Fprintln(log)
			res, err := Apply(plan, ApplyOptions{Version: version, SkipInstall: skipInstall, Out: log})
			if err != nil {
				return err
			}
			for _, s := range res.LostSlots {
				fmt.Fprintf(log, "pipeline slot %s holds steps the new pipeline has no place for: move them into another slot\n", s)
			}
			fmt.Fprintln(log, "\nConverted. Next: review `git status`, run `smith doctor` and `smith run lint typecheck test build`, then commit.")
			return nil
		},
	}
	f := cmd.Flags()
	f.BoolVar(&dryRun, "dry-run", false, "report the plan without changing anything")
	f.BoolVar(&asJSON, "json", false, "print the plan as JSON")
	f.BoolVar(&strict, "strict", false, "refuse to convert while anything needs manual work or is unsupported")
	f.BoolVar(&skipInstall, "skip-install", false, "do not run npm install afterwards")
	f.BoolVar(&allowDirty, "allow-dirty", false, "convert even with uncommitted changes")
	return cmd
}
