package release

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/MoNecromanCI/mnci/apps/mnci/session"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Command is `mnci release`.
func Command(version string) *cobra.Command {
	var (
		dryRun, yes, noPush, asJSON bool
	)
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Version, publish and tag every package with releasable commits",
		Long: `Release every publishable package from its conventional commits since its last tag:
feat → minor, fix/perf → patch, a breaking change → major (minor before 1.0).
Packages release in dependency order. Each is stamped, built, published and
tagged; manifests are restored afterwards (the tag is the record). A re-run
after a failure resumes where it stopped.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			plan, err := MakePlan(s.Git, s.Graph)
			if err != nil {
				return err
			}
			if !asJSON {
				printPlan(cmd, plan)
			}
			if len(plan.Items) == 0 {
				if asJSON {
					return json.NewEncoder(out).Encode(Report{DryRun: dryRun, Unchanged: plan.Unchanged, Results: []ReportResult{}})
				}
				return nil
			}
			if !dryRun && !yes {
				if !term.IsTerminal(int(os.Stdin.Fd())) {
					return errors.New("refusing to release without a terminal; pass --yes (CI does)")
				}
				fmt.Fprint(out, "Release these packages? [y/N] ")
				answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(answer), "y") {
					return errors.New("release cancelled")
				}
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
			defer stop()
			report, err := Execute(ctx, s, plan, Options{DryRun: dryRun, Push: !noPush, Out: out})
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				_ = enc.Encode(report)
			} else {
				fmt.Fprintln(out)
				fmt.Fprint(out, report.Markdown())
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be released; change nothing")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "do not ask for confirmation")
	cmd.Flags().BoolVar(&noPush, "no-push", false, "create tags locally only")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the release report as JSON")
	return cmd
}

func printPlan(cmd *cobra.Command, plan Plan) {
	out := cmd.OutOrStdout()
	if len(plan.Items) == 0 {
		fmt.Fprintln(out, "Nothing to release.")
	}
	for _, item := range plan.Items {
		from := item.Current
		if from == "" {
			from = "unreleased"
		}
		fmt.Fprintf(out, "  %-20s %s → %s  (%s)\n", item.Project, from, item.Next, item.Reason)
	}
	if len(plan.Unchanged) > 0 {
		fmt.Fprintf(out, "  unchanged: %s\n", strings.Join(plan.Unchanged, ", "))
	}
}
