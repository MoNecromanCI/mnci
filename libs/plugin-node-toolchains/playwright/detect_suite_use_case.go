package playwright

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// configFiles are the names the Playwright test runner reads its config from.
var configFiles = []string{"playwright.config.ts", "playwright.config.mts", "playwright.config.cts", "playwright.config.js", "playwright.config.mjs", "playwright.config.cjs"}

// Detect reads the project at dir (absolute) and says how it drives
// Playwright. A runner suite is recognised by its config or by
// @playwright/test; a library suite needs the playwright library and an e2e
// script to run, since nothing else says how its own runner is started.
func Detect(dir string, hasE2EScript bool) Suite {
	var m manifest
	if data, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
		_ = json.Unmarshal(data, &m)
	}
	s := Suite{Options: m.Smith.Playwright}
	switch {
	case s.Options.Flavour == Runner || s.Options.Flavour == Library:
		s.Flavour = s.Options.Flavour
	case hasConfig(dir) || m.declares("@playwright/test"):
		s.Flavour = Runner
	case (m.declares("playwright") || m.declares("playwright-core")) && hasE2EScript:
		s.Flavour = Library
	}
	if s.Flavour == Library && !hasE2EScript {
		s.Flavour = None
	}
	return s
}

func hasConfig(dir string) bool {
	for _, name := range configFiles {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	return false
}
