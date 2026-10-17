package declaredimports

import (
	"path"
	"strings"
)

// devFolders hold files that never ship: tests, end-to-end suites, test
// doubles (a slice's testing/ folder) and tools.
var devFolders = map[string]bool{"e2e": true, "tools": true, "test": true, "tests": true, "__tests__": true, "testing": true}

// DevFile reports whether a file, by its path relative to the project, is a
// test, e2e suite, tool or tool configuration: it may import what the
// package declares in devDependencies, and what the workspace root declares,
// since none of it is installed with the package.
func DevFile(rel string) bool {
	for _, dir := range strings.Split(path.Dir(rel), "/") {
		if devFolders[dir] {
			return true
		}
	}
	base := path.Base(rel)
	// jest.config.ts, vite.config.mts, playwright.config.ts, eslint.config.mjs…
	return strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") || strings.Contains(base, ".config.")
}
