package dependencies

import (
	"encoding/json"

	"github.com/MoNecromanCI/mnci/apps/mnci/session"
	"github.com/spf13/cobra"
)

// Commands are `mnci install`, `mnci uninstall` and `mnci sync`.
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
	return []*cobra.Command{install, uninstall, syncCommand(version)}
}

func syncCommand(version string) *cobra.Command {
	var check, asJSON, noInstall bool
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Make every project declare a shared dependency at the same version",
		Long: `Make every project declare a shared dependency at the same spec, language by
language. The target is the version installed, with the operator most
declarations use; when nothing says what is installed, the highest version
declared. Peer dependencies and the workspace's own packages are left alone,
and so is a spec mnci does not understand well enough to rewrite (a URL, a git
ref, a compound range): sync reports those. After rewriting it installs, so
lockfiles and the venv follow.`,
		Example: `  mnci sync --check   # in CI: fail when declarations disagree
  mnci sync`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			report, err := Sync(s, check, !noInstall)
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				if encErr := enc.Encode(report); encErr != nil {
					return encErr
				}
			} else {
				PrintSync(cmd.OutOrStdout(), report, check)
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "report only; fail when declarations disagree")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the report as JSON")
	cmd.Flags().BoolVar(&noInstall, "no-install", false, "rewrite the manifests without installing")
	return cmd
}
