package playwright

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Flavour is how a project drives Playwright.
type Flavour string

const (
	// None: the project has no Playwright suite this plugin knows how to run.
	None Flavour = ""
	// Runner: the Playwright test runner (@playwright/test, playwright.config.*).
	Runner Flavour = "runner"
	// Library: the playwright library driven by the project's own test runner
	// (Jest, Vitest, node --test) through its e2e script.
	Library Flavour = "library"
)

// configFiles are the names the Playwright test runner reads its config from.
var configFiles = []string{"playwright.config.ts", "playwright.config.mts", "playwright.config.cts", "playwright.config.js", "playwright.config.mjs", "playwright.config.cjs"}

// Options are what a project's package.json says under "smith": {"playwright": …}.
type Options struct {
	// Display: the suite opens headed windows (a recording flow, an Electron
	// or editor window), so on Linux it runs under xvfb-run.
	Display bool `json:"display"`
	// Browsers are what `playwright install` installs for it; default chromium.
	Browsers []string `json:"browsers"`
	// Flavour forces a flavour instead of detecting it.
	Flavour Flavour `json:"flavour"`
}

// manifest is the subset of package.json the plugin reads.
type manifest struct {
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	PeerDependencies     map[string]string `json:"peerDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
	Smith                struct {
		Playwright Options `json:"playwright"`
	} `json:"smith"`
}

func (m manifest) declares(name string) bool {
	for _, deps := range []map[string]string{m.Dependencies, m.DevDependencies, m.PeerDependencies, m.OptionalDependencies} {
		if _, ok := deps[name]; ok {
			return true
		}
	}
	return false
}

// Suite is a project's Playwright suite as the plugin sees it.
type Suite struct {
	Flavour Flavour
	Options Options
}

// BrowserList is the browsers to install, chromium when the project names none.
func (s Suite) BrowserList() []string {
	if len(s.Options.Browsers) == 0 {
		return []string{"chromium"}
	}
	return s.Options.Browsers
}

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
