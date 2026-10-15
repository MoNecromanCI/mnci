package pluginhost

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/MoNecromanCI/mnci/apps/mnci/workspace"
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
			ctor, ok := Builtin(args[0])
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
			for _, name := range sortedNames(BuiltinConventions) {
				rows = append(rows, row{name, serve.CapabilitiesOf(BuiltinConventions[name](version))})
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
	for _, enable := range []bool{true, false} {
		verb := map[bool]string{true: "enable", false: "disable"}[enable]
		cmd.AddCommand(&cobra.Command{
			Use:   verb + " <name>...",
			Short: strings.ToUpper(verb[:1]) + verb[1:] + " built-in plugins in mnci.toml",
			Args:  cobra.MinimumNArgs(1),
			RunE: func(c *cobra.Command, args []string) error {
				root, err := workspace.FindRoot(".")
				if err != nil {
					return err
				}
				cfg, err := workspace.Load(root)
				if err != nil {
					return err
				}
				enabled, err := Toggle(cfg.Plugins.Enabled, args, enable)
				if err != nil {
					return err
				}
				if err := workspace.SetEnabledPlugins(root, enabled); err != nil {
					return err
				}
				fmt.Fprintf(c.OutOrStdout(), "enabled plugins: %s\n", strings.Join(enabled, ", "))
				return nil
			},
		})
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "set <name>...",
		Short: "Enable exactly these built-in plugins in mnci.toml",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			root, err := workspace.FindRoot(".")
			if err != nil {
				return err
			}
			enabled, err := Toggle([]string{args[0]}, args, true)
			if err != nil {
				return err
			}
			if err := workspace.SetEnabledPlugins(root, enabled); err != nil {
				return err
			}
			fmt.Fprintf(c.OutOrStdout(), "enabled plugins: %s\n", strings.Join(enabled, ", "))
			return nil
		},
	})
	return cmd
}

// Toggle returns the enabled list after enabling or disabling names. An empty
// current list means every built-in plugin.
func Toggle(current, names []string, enable bool) ([]string, error) {
	if len(current) == 0 {
		current = BuiltinNames()
	}
	set := map[string]bool{}
	for _, n := range current {
		set[n] = true
	}
	for _, n := range names {
		if _, ok := Builtins[n]; !ok {
			return nil, fmt.Errorf("no built-in plugin %q (built-in: %s)", n, strings.Join(BuiltinNames(), ", "))
		}
		set[n] = enable
	}
	var out []string
	for _, n := range BuiltinNames() {
		if set[n] {
			out = append(out, n)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("at least one plugin must stay enabled")
	}
	return out, nil
}
