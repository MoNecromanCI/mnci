package graphview

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/MoNecromanCI/mnci/apps/mnci/affected"
	"github.com/MoNecromanCI/mnci/apps/mnci/session"
	"github.com/spf13/cobra"
)

// Commands returns `mnci graph`, `mnci projects` and `mnci affected`.
func Commands(version string) []*cobra.Command {
	var format, file string
	graph := &cobra.Command{
		Use:   "graph",
		Short: "Show how projects depend on each other",
		Example: `  mnci graph                       tree in the terminal
  mnci graph --format html -o graph.html
  mnci graph --format dot | dot -Tsvg > graph.svg`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			var out string
			switch format {
			case "text":
				out = Text(s.Graph)
			case "json":
				out = JSON(s.Graph)
			case "dot":
				out = DOT(s.Graph)
			case "html":
				out = HTML(s.Graph, s.Config.Workspace.Name)
			default:
				return fmt.Errorf("unknown format %q (text, json, dot, html)", format)
			}
			if file != "" {
				if err := os.WriteFile(file, []byte(out), 0o644); err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), "wrote "+file)
				return nil
			}
			fmt.Fprint(cmd.OutOrStdout(), out)
			return nil
		},
	}
	graph.Flags().StringVar(&format, "format", "text", "text, json, dot or html")
	graph.Flags().StringVarP(&file, "output", "o", "", "write to a file")

	var projectsJSON bool
	projects := &cobra.Command{
		Use:   "projects",
		Short: "List the workspace's projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			if projectsJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(Nodes(s.Graph))
			}
			for _, n := range Nodes(s.Graph) {
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-8s %-7s %s\n", n.Name, n.Kind, n.Language, n.Root)
			}
			return nil
		},
	}
	projects.Flags().BoolVar(&projectsJSON, "json", false, "print JSON")

	var base string
	var affectedJSON bool
	affectedCmd := &cobra.Command{
		Use:   "affected",
		Short: "List the projects affected by changes since a base branch",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			if base == "" {
				if base = affected.BaseFromEnv(); base == "" {
					base = s.Config.Release.Branch
				}
			}
			res := affected.Compute(s.Git, s.Graph, base)
			if affectedJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(res)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "%s\n", res.Reason)
			for _, p := range res.Projects {
				fmt.Fprintln(cmd.OutOrStdout(), p)
			}
			return nil
		},
	}
	affectedCmd.Flags().StringVar(&base, "base", "", "base branch (default: the PR target, else the release branch)")
	affectedCmd.Flags().BoolVar(&affectedJSON, "json", false, "print JSON")
	return []*cobra.Command{graph, projects, affectedCmd}
}
