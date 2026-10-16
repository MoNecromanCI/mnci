package dependencies

import (
	"github.com/MoNecromanCI/mnci/apps/mnci/session"
	"github.com/spf13/cobra"
)

// Commands are `mnci install` and `mnci uninstall`.
func Commands(version string) []*cobra.Command {
	var targets []string
	var dev bool
	install := &cobra.Command{
		Use:     "install [packages...]",
		Aliases: []string{"i"},
		Short:   "Add packages to a project with its own toolchain, or install the whole workspace",
		Long: `Add packages to one project with its own toolchain: npm install --workspace,
the project's pyproject.toml and the workspace venv, go get. Each project owns
its dependencies, so a package needs a target: mnci install -w <project>
<package>. With no package, mnci install installs everything a fresh clone
needs, for every language in the workspace.`,
		Example: `  mnci install                       # a fresh clone: npm install, the venv, go mod download
  mnci install -w web zod
  mnci install -w apps/api -D pytest-cov
  mnci install -w api github.com/pkg/errors@v0.9.1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			return Install(s, targets, args, dev, cmd.OutOrStdout())
		},
	}
	install.Flags().StringArrayVarP(&targets, "workspace", "w", nil, "target project, by name or directory; repeat it for several")
	install.Flags().BoolVarP(&dev, "save-dev", "D", false, "add as a development dependency (npm devDependencies, the Python dev extra, a Go tool)")

	var from []string
	uninstall := &cobra.Command{
		Use:     "uninstall <packages...>",
		Aliases: []string{"un"},
		Short:   "Remove packages from a project with its own toolchain",
		Long: `Remove packages from one project, the counterpart of mnci install. A
dependency belongs to the project that uses it, so it needs a target: mnci
uninstall -w <project> <package>.`,
		Example: `  mnci uninstall -w web zod`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			return Uninstall(s, from, args, cmd.OutOrStdout())
		},
	}
	uninstall.Flags().StringArrayVarP(&from, "workspace", "w", nil, "target project, by name or directory; repeat it for several")
	return []*cobra.Command{install, uninstall}
}
