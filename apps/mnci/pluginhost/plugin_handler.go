package pluginhost

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/serve"
	"github.com/spf13/cobra"
)

// Command is `mnci plugin`: list plugins, and serve a built-in one over stdio.
func Command(version string) *cobra.Command {
	cmd := &cobra.Command{Use: "plugin", Short: "Inspect plugins, or serve a built-in plugin"}
	cmd.AddCommand(&cobra.Command{
		Use:    "serve <name>",
		Short:  "Serve a built-in plugin on stdin/stdout (mnci runs this itself)",
		Args:   cobra.ExactArgs(1),
		Hidden: true,
		RunE: func(_ *cobra.Command, args []string) error {
			ctor, ok := Builtins[args[0]]
			if !ok {
				return fmt.Errorf("no built-in plugin %q", args[0])
			}
			return serve.Serve(ctor(version), os.Stdin, os.Stdout)
		},
	})
	var asJSON bool
	list := &cobra.Command{
		Use:   "list",
		Short: "List the built-in plugins and what each can do",
		RunE: func(cmd *cobra.Command, _ []string) error {
			type row struct {
				Name         string `json:"name"`
				Capabilities any    `json:"capabilities"`
			}
			var rows []row
			for _, name := range BuiltinNames() {
				rows = append(rows, row{name, serve.CapabilitiesOf(Builtins[name](version))})
			}
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(rows)
			}
			for _, r := range rows {
				fmt.Fprintf(cmd.OutOrStdout(), "%-8s %+v\n", r.Name, r.Capabilities)
			}
			return nil
		},
	}
	list.Flags().BoolVar(&asJSON, "json", false, "print JSON")
	cmd.AddCommand(list)
	return cmd
}
