package pluginhost

import (
	"fmt"
	"io"
	"os"
	"sort"
	"sync"

	"github.com/MoNecromanCI/mnci/apps/mnci/workspace"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/host"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Options configure Start.
type Options struct {
	WorkspaceRoot string
	Config        workspace.Config
	// Version is the host version, also handed to built-in plugins.
	Version string
	// Log receives plugin stderr.
	Log io.Writer
}

// Start launches every enabled plugin in parallel and initializes it.
func Start(opts Options) (*PluginSet, error) {
	self, err := os.Executable()
	if err != nil {
		return nil, err
	}
	type launch struct {
		name    string
		command []string
	}
	var launches []launch
	enabled := opts.Config.Plugins.Enabled
	if len(enabled) == 0 {
		enabled = BuiltinNames()
	}
	for _, name := range enabled {
		if _, ok := Builtins[name]; !ok {
			return nil, fmt.Errorf("mnci.toml enables unknown plugin %q; built-in plugins are %v", name, BuiltinNames())
		}
		if _, replaced := opts.Config.Plugins.External[name]; replaced {
			continue
		}
		launches = append(launches, launch{name, []string{self, "plugin", "serve", name}})
	}
	external := make([]string, 0, len(opts.Config.Plugins.External))
	for name := range opts.Config.Plugins.External {
		external = append(external, name)
	}
	sort.Strings(external)
	for _, name := range external {
		command, err := LaunchCommand(opts.WorkspaceRoot, opts.Config.Plugins.External, name)
		if err != nil {
			return nil, fmt.Errorf("mnci.toml [plugins.external.%s]: %w", name, err)
		}
		launches = append(launches, launch{name, command})
	}
	for _, name := range sortedNames(opts.Config.Conventions) {
		_, external := opts.Config.Plugins.External[name]
		_, builtin := BuiltinAddons[name]
		if !external && !builtin {
			return nil, fmt.Errorf("mnci.toml configures unknown convention %q; built-in ones are %v, and an external one is declared under [plugins.external.%s]", name, sortedNames(BuiltinAddons), name)
		}
	}
	for _, name := range sortedNames(BuiltinAddons) {
		if _, ok := opts.Config.Plugins.External[name]; ok {
			continue
		}
		launches = append(launches, launch{name, []string{self, "plugin", "serve", name}})
	}

	log := opts.Log
	if log == nil {
		log = os.Stderr
	}
	clients := make([]*host.Client, len(launches))
	errs := make([]error, len(launches))
	var wg sync.WaitGroup
	for i, l := range launches {
		wg.Add(1)
		go func() {
			defer wg.Done()
			clients[i], errs[i] = host.Start(host.Options{
				Command:       l.command,
				Env:           append(os.Environ(), "MNCI_PLUGIN=1"),
				WorkspaceRoot: opts.WorkspaceRoot,
				Stderr:        log,
				Host:          protocol.Peer{Name: "mnci", Version: opts.Version},
			})
		}()
	}
	wg.Wait()
	set := &PluginSet{}
	for i, c := range clients {
		if errs[i] != nil {
			set.Clients = append(set.Clients, clients[:i]...)
			set.Close()
			return nil, fmt.Errorf("plugin %s: %w", launches[i].name, errs[i])
		}
		set.Clients = append(set.Clients, c)
	}
	for _, name := range external {
		if err := checkExternal(set, name, opts.Config.Plugins.External[name]); err != nil {
			set.Close()
			return nil, err
		}
	}
	return set, nil
}

// checkExternal checks that an external plugin answered with the name it is
// declared under and a version mnci.toml accepts.
func checkExternal(set *PluginSet, name string, cfg workspace.ExternalPlugin) error {
	c := set.Get(name)
	if c == nil {
		return fmt.Errorf("mnci.toml [plugins.external.%s]: the plugin calls itself something else; declare it under its own name", name)
	}
	ok, err := Satisfies(c.Info.Plugin.Version, cfg.Version)
	if err != nil {
		return fmt.Errorf("mnci.toml [plugins.external.%s]: %w", name, err)
	}
	if !ok {
		return fmt.Errorf("mnci.toml [plugins.external.%s] wants version %s, but the plugin is %s", name, cfg.Version, c.Info.Plugin.Version)
	}
	return nil
}
