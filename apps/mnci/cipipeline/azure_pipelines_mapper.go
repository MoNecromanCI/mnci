package cipipeline

import (
	"fmt"
	"strings"
)

// AzurePath is where the Azure Pipelines definition is written.
const AzurePath = "azure-pipelines.yml"

var azureSetup = map[string]string{
	"node":   "  - task: UseNode@1\n    inputs:\n      version: '%s.x'\n",
	"python": "  - task: UsePythonVersion@0\n    inputs:\n      versionSpec: '%s'\n",
	"go":     "  - task: GoTool@0\n    inputs:\n      version: '%s'\n",
}

// RenderAzure renders the model as an Azure Pipelines definition.
func RenderAzure(m Model) string {
	var b strings.Builder
	b.WriteString(header("#"))
	fmt.Fprintf(&b, `trigger:
  branches:
    include: [%[1]s]

pr:
  branches:
    include: [%[1]s]

pool:
  vmImage: ubuntu-latest

steps:
  - checkout: self
    fetchDepth: 0
    persistCredentials: true
`, m.ReleaseBranch)
	for _, t := range m.Toolchains {
		if tmpl, ok := azureSetup[t.Name]; ok {
			version := t.Version
			if t.Name == "go" && strings.Count(version, ".") == 1 {
				version += ".0"
			}
			fmt.Fprintf(&b, tmpl, version)
		}
	}
	fmt.Fprintf(&b, `  - bash: curl -fsSL %s | sh -s -- "%s"
    displayName: Install mnci
`, InstallScriptURL, m.MnciVersion)
	b.WriteString(slot("  ", "before-setup"))
	b.WriteString(`  - bash: mnci ci setup
    displayName: Setup
  - bash: mnci ci audit
    displayName: Audit
  - bash: mnci ci verify
    displayName: Verify
`)
	b.WriteString(slot("  ", "after-verify"))
	fmt.Fprintf(&b, `  - bash: mnci release --dry-run
    displayName: Release preview
    condition: and(succeeded(), eq(variables['Build.Reason'], 'PullRequest'))
  - bash: mnci ci release
    displayName: Release
    condition: and(succeeded(), in(variables['Build.Reason'], 'IndividualCI', 'BatchedCI'), eq(variables['Build.SourceBranch'], 'refs/heads/%s'))
`, m.ReleaseBranch)
	if len(m.Credentials) > 0 {
		b.WriteString("    env:\n")
		for _, c := range m.Credentials {
			fmt.Fprintf(&b, "      %s: $(%s)\n", c.Env, c.Env)
		}
	}
	return b.String()
}
