package vitest

import (
	"github.com/MoNecromanCI/mnci/libs/plugin-node-conventions/tooling"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Name is the convention's name in mnci.toml ([conventions.vitest]).
const Name = "vitest"

// New returns the convention; version is reported in the handshake.
func New(version string) *tooling.Convention {
	return &tooling.Convention{Name: Name, Version: version, Setup: setup}
}

func setup(workspaceRoot string, params protocol.GenerateParams) tooling.Setup {
	return tooling.Setup{
		DevDependencies: map[string]string{"vitest": "^5.0.5"},
		Files:           tooling.StarterSpec(workspaceRoot, params, "vitest"),
		Scripts:         map[string]string{"test": "vitest run"},
	}
}
