package pluginhost

import (
	"sort"

	dotnetp "github.com/MoNecromanCI/mnci/libs/plugin-dotnet/plugin"
	gop "github.com/MoNecromanCI/mnci/libs/plugin-go/plugin"
	"github.com/MoNecromanCI/mnci/libs/plugin-go/pluginkit"
	"github.com/MoNecromanCI/mnci/libs/plugin-node-toolchains/eslint"
	"github.com/MoNecromanCI/mnci/libs/plugin-node-toolchains/jest"
	"github.com/MoNecromanCI/mnci/libs/plugin-node-toolchains/typescript"
	"github.com/MoNecromanCI/mnci/libs/plugin-node-toolchains/vitest"
	nodep "github.com/MoNecromanCI/mnci/libs/plugin-node/plugin"
	pythonp "github.com/MoNecromanCI/mnci/libs/plugin-python/plugin"
	rustp "github.com/MoNecromanCI/mnci/libs/plugin-rust/plugin"
	slicesp "github.com/MoNecromanCI/mnci/libs/plugin-vertical-slices/plugin"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/serve"
)

// Builtins constructs each built-in language plugin by name.
var Builtins = map[string]func(version string) serve.Plugin{
	"node":   func(v string) serve.Plugin { return nodep.New(v) },
	"go":     func(v string) serve.Plugin { return gop.New(v) },
	"dotnet": func(v string) serve.Plugin { return dotnetp.New(v) },
	"python": func(v string) serve.Plugin { return pythonp.New(v) },
	"rust":   func(v string) serve.Plugin { return rustp.New(v) },
}

// BuiltinAddons constructs each built-in plugin that is not a language:
// toolchain, structure, scaffold and deploy plugins. They are always loaded,
// so their stacks are available, but act only on projects whose stack names
// them, or as a convention mnci.toml configures ([conventions.<name>]).
var BuiltinAddons = map[string]func(version string) serve.Plugin{
	slicesp.Name:    func(v string) serve.Plugin { return slicesp.New(v) },
	eslint.Name:     func(v string) serve.Plugin { return eslint.New(v) },
	jest.Name:       func(v string) serve.Plugin { return jest.New(v) },
	vitest.Name:     func(v string) serve.Plugin { return vitest.New(v) },
	typescript.Name: func(v string) serve.Plugin { return typescript.New(v) },
	pluginkit.Name:  func(v string) serve.Plugin { return pluginkit.New(v) },
}

// Builtin returns the constructor of a built-in plugin.
func Builtin(name string) (func(version string) serve.Plugin, bool) {
	if ctor, ok := Builtins[name]; ok {
		return ctor, true
	}
	ctor, ok := BuiltinAddons[name]
	return ctor, ok
}

// BuiltinNames lists the built-in language plugins, sorted.
func BuiltinNames() []string { return sortedNames(Builtins) }

func sortedNames[V any](m map[string]V) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
