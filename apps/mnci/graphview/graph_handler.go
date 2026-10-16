package graphview

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/MoNecromanCI/mnci/apps/mnci/affected"
	"github.com/MoNecromanCI/mnci/apps/mnci/session"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// Commands returns `mnci graph`, `mnci projects` and `mnci affected`.
func Commands(version string) []*cobra.Command {
	var format, file, graphBase string
	var focus []string
	var onlyAffected bool
	graph := &cobra.Command{
		Use:   "graph",
		Short: "Show how projects depend on each other",
		Long: `Show the project graph as a tree (text), JSON, Graphviz (dot), a static page
(html) or an interactive browser (tui). Every format shows the same projects:
--focus keeps the named projects and everything they depend on or that depends
on them; --affected keeps the projects affected by changes since --base and
marks the ones that changed.`,
		Example: `  mnci graph                       tree in the terminal
  mnci graph --format tui          browse it
  mnci graph --focus web           web, what it uses and what uses it
  mnci graph --affected --format html -o affected.html
  mnci graph --format dot | dot -Tsvg > graph.svg`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			g, marked := s.Graph, map[string]bool{}
			if len(focus) > 0 {
				names, err := s.Graph.Select(focus)
				if err != nil {
					return err
				}
				for _, n := range names {
					marked[n] = true
				}
				g = Around(s.Graph, names)
			}
			if onlyAffected {
				base := graphBase
				if base == "" {
					if base = affected.BaseFromEnv(); base == "" {
						base = s.Config.Release.Branch
					}
				}
				res := affected.Compute(s.Git, s.Graph, base)
				fmt.Fprintln(cmd.ErrOrStderr(), "affected: "+res.Reason)
				keep := map[string]bool{}
				for _, n := range res.Projects {
					if g.Get(n) != nil {
						keep[n] = true
					}
				}
				for _, f := range res.Changed {
					if owner := s.Graph.Owner(f); owner != nil {
						marked[owner.Name] = true
					}
				}
				g = Subgraph(g, keep)
			}
			var out string
			switch format {
			case "text":
				out = Text(g, marked)
			case "json":
				out = JSON(g, marked)
			case "dot":
				out = DOT(g, marked)
			case "html":
				out = HTML(g, s.Config.Workspace.Name, marked)
			case "tui":
				_, err := tea.NewProgram(NewBrowser(Nodes(g, marked)), tea.WithOutput(cmd.OutOrStdout())).Run()
				return err
			default:
				return fmt.Errorf("unknown format %q (text, json, dot, html, tui)", format)
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
	graph.Flags().StringVar(&format, "format", "text", "text, json, dot, html or tui")
	graph.Flags().StringVarP(&file, "output", "o", "", "write to a file")
	graph.Flags().StringSliceVar(&focus, "focus", nil, "keep these projects and everything connected to them (names, globs, tag:x)")
	graph.Flags().BoolVar(&onlyAffected, "affected", false, "keep the projects affected by changes since --base; mark the changed ones")
	graph.Flags().StringVar(&graphBase, "base", "", "base branch for --affected (default: the PR target, else the release branch)")

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
				return enc.Encode(Nodes(s.Graph, nil))
			}
			for _, n := range Nodes(s.Graph, nil) {
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
