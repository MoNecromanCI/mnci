package scaffold

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/MoNecromanCI/mnci/apps/mnci/session"
	"github.com/spf13/cobra"
)

// NewCommand is `mnci new`.
func NewCommand(version string) *cobra.Command {
	var opts NewOptions
	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a monorepo: configuration, git, CI pipelines and a first commit",
		Example: `  mnci new acme
  mnci new acme --dir . --ci github,azure   # in an existing clone`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name, opts.Version, opts.Out = args[0], version, cmd.OutOrStdout()
			dir, err := NewWorkspace(opts)
			if err != nil {
				return err
			}
			cwd, _ := filepath.Abs(".")
			rel, err := filepath.Rel(cwd, dir)
			if err != nil {
				rel = dir
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nCreated %s. Next:\n  cd %s\n  mnci add app web --lang node\n", opts.Name, rel)
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.Dir, "dir", "", "directory to create the workspace in (default ./<name>)")
	cmd.Flags().StringVar(&opts.Scope, "scope", "", "npm scope for packages (default @<name>)")
	cmd.Flags().StringVar(&opts.ModulePrefix, "module-prefix", "", "Go module prefix (default from the origin remote, else example.com/<name>)")
	cmd.Flags().StringSliceVar(&opts.Providers, "ci", nil, "pipelines to generate: github, azure (default github)")
	return cmd
}

// AddCommand is `mnci add`.
func AddCommand(version string) *cobra.Command {
	var lang, stack string
	cmd := &cobra.Command{
		Use:   "add <kind> <name>",
		Short: "Add an app, an internal lib or a publishable package",
		Long: `Add a project. The kind says what happens to it:

  app      deployed, never published         → apps/<name>
  lib      imported internally, never published → libs/<name>
  package  published to its registry on release → packages/<name>

A combined kind from ` + "`mnci kinds`" + ` names the language too: node-app is
app --lang node.

The project lints, type-checks, tests and builds immediately, and the CI
pipelines are regenerated to install its toolchain.`,
		Example: `  mnci add app web --lang node
  mnci add app api --stack node-vitest
  mnci add lib core --lang go
  mnci add package sdk --lang python`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			kind := args[0]
			// "node-app" is "app --lang node": the form `mnci kinds` lists.
			if k, l, ok := SplitKind(kind); ok && lang == "" {
				kind, lang = k, l
			}
			res, err := AddProject(s, AddOptions{Kind: kind, Name: args[1], Lang: lang, Stack: stack, Out: cmd.OutOrStdout()})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Wrote %d file(s):\n  %s\n", len(res.Files), strings.Join(res.Files, "\n  "))
			for _, n := range res.Notes {
				fmt.Fprintln(out, n)
			}
			fmt.Fprintf(out, "Try: mnci test -p %s\n", args[1])
			return nil
		},
	}
	cmd.Flags().StringVarP(&lang, "lang", "l", "", "language: node, go, python, rust, dotnet, ruby, php, cpp, dart, jvm (default: the stack's, or the workspace's only language)")
	cmd.Flags().StringVarP(&stack, "stack", "s", "", "stack the project is built with (default: [default_stacks] in mnci.toml; mnci stacks lists them; none: the language plugin alone)")
	return cmd
}
