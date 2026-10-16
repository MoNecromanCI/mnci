package cipipeline

import (
	"fmt"
	"strings"
)

// GitHubSetup is how every job mnci writes starts, as GitHub Actions steps
// (at a job's steps indentation): checkout with full history, the
// toolchains, mnci at the version that wrote it, and mnci ci setup.
func GitHubSetup(m Model) string {
	var b strings.Builder
	b.WriteString(`      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
`)
	for _, t := range m.Toolchains {
		b.WriteString(githubSetup(t))
	}
	fmt.Fprintf(&b, `      - name: Install mnci
        run: curl -fsSL %s | sh -s -- "$MNCI_VERSION"
        env:
          MNCI_VERSION: %s
          GITHUB_TOKEN: ${{ github.token }}
      - name: Setup
        run: mnci ci setup
`, InstallScriptURL, m.MnciVersion)
	return b.String()
}

// AzureSetup is the same start as Azure Pipelines steps (at the top level's
// indentation).
func AzureSetup(m Model) string {
	var b strings.Builder
	b.WriteString(`  - checkout: self
    fetchDepth: 0
`)
	for _, t := range m.Toolchains {
		b.WriteString(azureSetup(t))
	}
	fmt.Fprintf(&b, `  - bash: curl -fsSL %s | sh -s -- "%s"
    displayName: Install mnci
  - bash: mnci ci setup
    displayName: Setup
`, InstallScriptURL, m.MnciVersion)
	return b.String()
}
