package jest

import (
	"github.com/MoNecromanCI/mnci/libs/plugin-node-toolchains/tooling"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Name is the plugin's name in a stack's toolchain list.
const Name = "jest"

// stacks are the stacks this plugin publishes.
var stacks = []protocol.Stack{{
	Name:        "node-jest",
	Description: "TypeScript on Node, linted by @mnci/eslint-config, tested with Jest",
	Language:    "node",
	Toolchain:   []string{"eslint", Name},
}}

// New returns the plugin; version is reported in the handshake.
func New(version string) *tooling.Toolchain {
	return &tooling.Toolchain{Name: Name, Version: version, Provides: []string{"test"}, Stacks: stacks, Setup: setup}
}

func setup(workspaceRoot string, params protocol.GenerateParams) tooling.Setup {
	return tooling.Setup{
		DevDependencies: map[string]string{
			"@jest/globals": "^30.5.2",
			"jest":          "^30.5.2",
			"ts-jest":       "^29.4.15",
		},
		Files:   tooling.StarterSpec(workspaceRoot, params, "@jest/globals"),
		Scripts: map[string]string{"test": "jest"},
		Fields: map[string]any{"jest": map[string]any{
			"preset":          "ts-jest",
			"testEnvironment": "node",
			"roots":           []string{"<rootDir>/src"},
		}},
	}
}
