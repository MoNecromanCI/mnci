package eslint

import (
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tooling"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the plugin's name in a stack's toolchain list.
const Name = "eslint"

const config = `// Lint AND formatting for every Node project: @monosmith/eslint-config carries the
// TypeScript rules, JavaScript Standard Style and the vertical-slice rules.
// Add your own blocks after the spread; later blocks win.
import monosmith from '@monosmith/eslint-config'

export default [
  ...monosmith(),
]
`

// New returns the plugin; version is reported in the handshake.
func New(version string) *tooling.Toolchain {
	return &tooling.Toolchain{Name: Name, Version: version, Provides: []string{"lint"}, Setup: setup}
}

func setup(string, protocol.GenerateParams) tooling.Setup {
	return tooling.Setup{
		RootFiles: map[string]string{"eslint.config.mjs": config},
		DevDependencies: map[string]string{
			"@monosmith/eslint-config": "^1.0.0",
			"eslint":                   "^10.12.0",
		},
		Scripts: map[string]string{"lint": "eslint ."},
	}
}
