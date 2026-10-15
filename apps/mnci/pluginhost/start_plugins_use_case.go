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
		launches = append(launches, launch{name, []string{self, "plugin", "serve", name}})
	}
	external := make([]string, 0, len(opts.Config.Plugins.External))
	for name := range opts.Config.Plugins.External {
		external = append(external, name)
	}
	sort.Strings(external)
	for _, name := range external {
		launches = append(launches, launch{name, opts.Config.Plugins.External[name].Command})
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
	return set, nil
}
