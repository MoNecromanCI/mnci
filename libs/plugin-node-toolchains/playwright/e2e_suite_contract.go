package playwright

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

// declaresLibrary reports whether the manifest declares the playwright
// library (or playwright-core, which it is built on).
func (m manifest) declaresLibrary() bool {
	return m.declares("playwright") || m.declares("playwright-core")
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
