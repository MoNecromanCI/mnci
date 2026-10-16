package explain

import (
	"encoding/json"
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// Command is `smith explain`.
func Command(version string) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "explain <project>[:<task>]",
		Short:   "Show how a project's tasks resolved, and where each setting came from",
		Example: "  smith explain web\n  smith explain web:build",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			tasks, err := Explain(s.Graph, args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(tasks)
			}
			for _, t := range tasks {
				fmt.Fprintf(out, "%s", t.ID)
				if t.Description != "" {
					fmt.Fprintf(out, " — %s", t.Description)
				}
				fmt.Fprintln(out)
				for _, f := range t.Fields {
					fmt.Fprintf(out, "  %-10s %s\n  %-10s   from %s\n", f.Name, f.Value, "", f.Source)
				}
				fmt.Fprintln(out)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON")
	return cmd
}
