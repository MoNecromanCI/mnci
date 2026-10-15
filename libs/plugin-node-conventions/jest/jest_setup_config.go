package jest

import (
	"github.com/MoNecromanCI/mnci/libs/plugin-node-conventions/tooling"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Name is the convention's name in mnci.toml ([conventions.jest]).
const Name = "jest"

// New returns the convention; version is reported in the handshake.
func New(version string) *tooling.Convention {
	return &tooling.Convention{Name: Name, Version: version, Setup: setup}
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
