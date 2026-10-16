package cipipeline

import (
	"fmt"
	"strings"
)

// AzurePath is where the Azure Pipelines definition is written.
const AzurePath = "azure-pipelines.yml"

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
		b.WriteString(azureSetup(t))
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
	b.WriteString(`  - bash: mnci release --dry-run --comment
    displayName: Release preview
    condition: and(succeeded(), eq(variables['Build.Reason'], 'PullRequest'))
    env:
      SYSTEM_ACCESSTOKEN: $(System.AccessToken)
`)
	onPush := fmt.Sprintf("and(succeeded(), in(variables['Build.Reason'], 'IndividualCI', 'BatchedCI'), eq(variables['Build.SourceBranch'], 'refs/heads/%s'))", m.ReleaseBranch)
	if m.Pack {
		fmt.Fprintf(&b, `  - bash: mnci ci pack
    displayName: Pack
    condition: %[1]s
  - publish: dist/drop
    artifact: drop
    displayName: Keep the packed apps
    condition: %[1]s
`, onPush)
	}
	fmt.Fprintf(&b, `  - bash: mnci ci release
    displayName: Release
    condition: %s
`, onPush)
	if len(m.Credentials) > 0 {
		b.WriteString("    env:\n")
		for _, c := range m.Credentials {
			fmt.Fprintf(&b, "      %s: $(%s)\n", c.Env, c.Env)
		}
	}
	return b.String()
}
