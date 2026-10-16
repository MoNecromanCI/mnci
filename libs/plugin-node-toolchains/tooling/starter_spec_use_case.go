package tooling

import (
	"os"
	"path/filepath"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

const starterUseCase = "src/greeting/greet.use-case.ts"

// StarterSpec returns the spec of the greeting slice the node plugin starts
// every project with, importing the test API from module ("@jest/globals",
// "vitest"). A project without that slice gets no spec.
func StarterSpec(workspaceRoot string, params protocol.GenerateParams, module string) map[string]string {
	if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory), filepath.FromSlash(starterUseCase))); err != nil {
		return nil
	}
	return map[string]string{"src/greeting/greet.use-case.spec.ts": `import { describe, expect, it } from '` + module + `'
import { greet } from './greet.use-case.ts'

describe('greet', () => {
  it('greets by name', () => {
    expect(greet('world').message).toBe('Hello, world!')
  })
})
`}
}
