package pluginhost

import (
	"sort"

	gop "github.com/MoNecromanCI/mnci/libs/plugin-go/plugin"
	nodep "github.com/MoNecromanCI/mnci/libs/plugin-node/plugin"
	pythonp "github.com/MoNecromanCI/mnci/libs/plugin-python/plugin"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/serve"
)

// Builtins constructs each built-in plugin by name.
var Builtins = map[string]func(version string) serve.Plugin{
	"node":   func(v string) serve.Plugin { return nodep.New(v) },
	"go":     func(v string) serve.Plugin { return gop.New(v) },
	"python": func(v string) serve.Plugin { return pythonp.New(v) },
}

// BuiltinNames lists the built-in plugins, sorted.
func BuiltinNames() []string {
	names := make([]string, 0, len(Builtins))
	for name := range Builtins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
