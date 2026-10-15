package eslint

import (
	"github.com/MoNecromanCI/mnci/libs/plugin-node-toolchains/tooling"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Name is the convention's name in mnci.toml ([conventions.eslint]).
const Name = "eslint"

const config = `// Lint AND formatting for every Node project: @mnci/eslint-config carries the
// TypeScript rules, JavaScript Standard Style and the vertical-slice rules.
// Add your own blocks after the spread; later blocks win.
import mnci from '@mnci/eslint-config'

export default [
  ...mnci(),
]
`

// New returns the convention; version is reported in the handshake.
func New(version string) *tooling.Convention {
	return &tooling.Convention{Name: Name, Version: version, Setup: setup}
}

func setup(string, protocol.GenerateParams) tooling.Setup {
	return tooling.Setup{
		RootFiles: map[string]string{"eslint.config.mjs": config},
		DevDependencies: map[string]string{
			"@mnci/eslint-config": "^0.3.31",
			"eslint":              "^10.12.0",
		},
		Scripts: map[string]string{"lint": "eslint ."},
	}
}
