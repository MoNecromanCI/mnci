package cipipeline

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/MoNecromanCI/mnci/apps/mnci/session"
	"github.com/spf13/cobra"
)

// Command is `mnci ci`.
func Command(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ci",
		Short: "Run a CI phase, or regenerate the pipelines",
		Long: `The generated pipelines call these phases, so CI runs exactly what you can run locally:

  setup    install every toolchain dependency the workspace's plugins need
  audit    check dependencies for known vulnerabilities
  verify   lint, typecheck, test and build — only affected projects on a pull request
  release  version, publish and tag every package with releasable commits`,
	}
	for _, phase := range Phases {
		cmd.AddCommand(&cobra.Command{
			Use:   phase,
			Short: "Run the " + phase + " phase",
			Args:  cobra.NoArgs,
			RunE: func(c *cobra.Command, _ []string) error {
				s, err := session.Open(session.Options{Dir: ".", Version: version, Log: c.ErrOrStderr()})
				if err != nil {
					return err
				}
				defer s.Close()
				ctx, stop := signal.NotifyContext(c.Context(), os.Interrupt)
				defer stop()
				return RunPhase(ctx, s, phase, c.OutOrStdout())
			},
		})
	}
	var check bool
	generate := &cobra.Command{
		Use:   "generate",
		Short: "Rewrite the CI pipelines from the workspace's projects (--check only reports drift)",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: c.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			if check {
				contributions, err := Collect(s)
				if err != nil {
					return err
				}
				if drift := Drift(s.Root, s.Config, NewModel(contributions, s.Version, s.Config.Release.Branch)); len(drift) > 0 {
					return errors.New("pipelines are out of date: " + strings.Join(drift, "; ") + " — run `mnci ci generate`")
				}
				fmt.Fprintln(c.OutOrStdout(), "pipelines are up to date")
				return nil
			}
			changed, lost, err := Regenerate(s)
			if err != nil {
				return err
			}
			for _, f := range changed {
				fmt.Fprintln(c.OutOrStdout(), "wrote "+f)
			}
			for _, l := range lost {
				fmt.Fprintln(c.ErrOrStderr(), "warning: slot no longer exists, its content was not kept: "+l)
			}
			if len(changed) == 0 {
				fmt.Fprintln(c.OutOrStdout(), "pipelines are up to date")
			}
			return nil
		},
	}
	generate.Flags().BoolVar(&check, "check", false, "report drift without writing")
	cmd.AddCommand(generate)
	return cmd
}
