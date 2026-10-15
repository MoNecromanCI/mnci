package slicecheck

import (
	"fmt"

	"github.com/MoNecromanCI/mnci/libs/plugin-vertical-slices/layout"
	"github.com/spf13/cobra"
)

// Command is `mnci slices`.
func Command() *cobra.Command {
	cmd := &cobra.Command{Use: "slices", Short: "Check the vertical-slice layout of projects"}
	var lang string
	check := &cobra.Command{
		Use:   "check <project directory>...",
		Short: "Check that files are role-named and no folder is a junk drawer",
		Long: `Checks Go or Python projects against the vertical-slice rules: a file is
<name>_<role> (roles: ` + fmt.Sprint(layout.Roles) + `),
a module's or package's root holds only its entry files, and no folder is
named util, helper, common, shared, manager, processor, data or misc. A slice
with more than ` + fmt.Sprint(layout.SliceSize) + ` production files is reported for review.

The vertical-slices convention ([conventions.vertical-slices] in mnci.toml)
runs this as each project's slice-check task.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, dirs []string) error {
			problems := 0
			for _, dir := range dirs {
				r, err := layout.Check(dir, lang)
				if err != nil {
					return err
				}
				for _, w := range r.Warnings {
					fmt.Fprintln(cmd.OutOrStdout(), "review: "+w)
				}
				for _, p := range r.Problems {
					fmt.Fprintln(cmd.ErrOrStderr(), p)
				}
				problems += len(r.Problems)
			}
			if problems > 0 {
				return fmt.Errorf("%d file(s) or folder(s) break the vertical-slice layout; an exception is written in the project README", problems)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "vertical-slice layout OK")
			return nil
		},
	}
	check.Flags().StringVar(&lang, "lang", "go", "language of the projects: go or python")
	cmd.AddCommand(check)
	return cmd
}
