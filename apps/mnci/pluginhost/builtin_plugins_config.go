package pluginhost

import (
	"sort"

	gop "github.com/MoNecromanCI/mnci/libs/plugin-go/plugin"
	"github.com/MoNecromanCI/mnci/libs/plugin-node-toolchains/eslint"
	"github.com/MoNecromanCI/mnci/libs/plugin-node-toolchains/jest"
	"github.com/MoNecromanCI/mnci/libs/plugin-node-toolchains/vitest"
	nodep "github.com/MoNecromanCI/mnci/libs/plugin-node/plugin"
	pythonp "github.com/MoNecromanCI/mnci/libs/plugin-python/plugin"
	slicesp "github.com/MoNecromanCI/mnci/libs/plugin-vertical-slices/plugin"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/serve"
)

// Builtins constructs each built-in plugin by name.
var Builtins = map[string]func(version string) serve.Plugin{
	"node":   func(v string) serve.Plugin { return nodep.New(v) },
	"go":     func(v string) serve.Plugin { return gop.New(v) },
	"python": func(v string) serve.Plugin { return pythonp.New(v) },
}

// BuiltinConventions constructs each built-in convention plugin by name. A
// convention runs only in a workspace whose mnci.toml configures it
// ([conventions.<name>]).
var BuiltinConventions = map[string]func(version string) serve.Plugin{
	slicesp.Name: func(v string) serve.Plugin { return slicesp.New(v) },
	eslint.Name:  func(v string) serve.Plugin { return eslint.New(v) },
	jest.Name:    func(v string) serve.Plugin { return jest.New(v) },
	vitest.Name:  func(v string) serve.Plugin { return vitest.New(v) },
}

// Builtin returns the constructor of a built-in plugin or convention.
func Builtin(name string) (func(version string) serve.Plugin, bool) {
	if ctor, ok := Builtins[name]; ok {
		return ctor, true
	}
	ctor, ok := BuiltinConventions[name]
	return ctor, ok
}

// BuiltinNames lists the built-in plugins, sorted.
func BuiltinNames() []string { return sortedNames(Builtins) }

func sortedNames[V any](m map[string]V) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
