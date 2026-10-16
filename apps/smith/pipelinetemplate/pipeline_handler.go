package pipelinetemplate

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/MoNecromanCI/mnci/apps/mnci/session"
	"github.com/spf13/cobra"
)

// Command is `mnci pipeline`.
func Command(version string) *cobra.Command {
	var opts Options
	var ci string
	cmd := &cobra.Command{
		Use:   "pipeline [template]",
		Short: "Add a ready-made pipeline: e2e, package-zip, deploy-pages, deploy-azure-function",
		Long: `Add a ready-made pipeline file to the workspace, beside the generated CI: it
starts the way every generated job does (the toolchains, mnci, mnci ci setup),
then does one thing. It is written once and yours from then on: mnci upgrade
never rewrites it, and an existing file is kept unless --force. With no
template, lists them.`,
		Example: `  mnci pipeline
  mnci pipeline e2e
  mnci pipeline deploy-pages --project web`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if len(args) == 0 {
				w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
				for _, t := range Catalog {
					needs := ""
					if t.NeedsProject {
						needs = " (--project <app>)"
					}
					fmt.Fprintf(w, "%s%s\t%s\t%s\n", t.Name, needs, strings.Join(t.Providers, ", "), t.Description)
				}
				return w.Flush()
			}
			switch ci {
			case "":
			case "both":
				opts.Providers = []string{"github", "azure"}
			case "github", "azure":
				opts.Providers = []string{ci}
			default:
				return fmt.Errorf("--ci is github, azure or both, not %q", ci)
			}
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			written, err := Add(s, args[0], opts)
			for _, f := range written {
				fmt.Fprintf(out, "✔ wrote %s\n", f)
			}
			return err
		},
	}
	cmd.Flags().StringVar(&opts.Project, "project", "", "the app a deploy template builds")
	cmd.Flags().StringVar(&ci, "ci", "", "github, azure or both (default: the workspace's providers)")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "replace a pipeline file that exists")
	return cmd
}
