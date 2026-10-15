package doctor

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/MoNecromanCI/mnci/apps/mnci/session"
	"github.com/spf13/cobra"
)

// ErrUnhealthy makes doctor exit non-zero.
var ErrUnhealthy = errors.New("the workspace has problems")

// Command is `mnci doctor`.
func Command(version string) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check the workspace's health; every problem names its remedy",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			report := Run(s)
			out := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				_ = enc.Encode(report)
			} else {
				for _, f := range report.Findings {
					mark := "✔"
					if !f.OK {
						mark = "✖"
						if f.Warning {
							mark = "!"
						}
					}
					fmt.Fprintf(out, "%s %s\n", mark, f.Check)
					if !f.OK {
						fmt.Fprintf(out, "    %s\n    → %s\n", f.Detail, f.Remedy)
					}
				}
				fmt.Fprintf(out, "\n%d passed, %d failed, %d warning(s)\n", report.Passed, report.Failed, report.Warnings)
			}
			if report.Failed > 0 {
				return ErrUnhealthy
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the report as JSON")
	return cmd
}
