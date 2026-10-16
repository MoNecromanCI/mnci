package cipipeline

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// githubSetup renders the steps that install a toolchain on GitHub Actions:
// the plugin's action, else its install command, else a comment saying the
// plugin gave neither.
func githubSetup(t protocol.Toolchain) string {
	switch {
	case t.GitHub != nil:
		return actionStep("      ", "- uses: ", "with", t.GitHub)
	case t.Install != "":
		return fmt.Sprintf("      - name: Install %s\n        run: %s", t.Name, runValue("        ", t.Install))
	}
	return missing("      ", t)
}

// azureSetup renders the steps that install a toolchain on Azure Pipelines.
func azureSetup(t protocol.Toolchain) string {
	switch {
	case t.Azure != nil:
		return actionStep("  ", "- task: ", "inputs", t.Azure)
	case t.Install != "":
		return fmt.Sprintf("  - bash: %s    displayName: Install %s\n", runValue("    ", t.Install), t.Name)
	}
	return missing("  ", t)
}

func actionStep(indent, head, inputsKey string, s *protocol.ActionStep) string {
	var b strings.Builder
	b.WriteString(indent + head + s.Uses + "\n")
	if len(s.Inputs) > 0 {
		b.WriteString(indent + "  " + inputsKey + ":\n")
		for _, in := range s.Inputs {
			b.WriteString(indent + "    " + in.Name + ": " + yamlValue(in.Value) + "\n")
		}
	}
	return b.String()
}

// runValue is a run/bash value: inline for one line, a block for several.
func runValue(indent, command string) string {
	command = strings.TrimRight(command, "\n")
	if !strings.Contains(command, "\n") {
		return command + "\n"
	}
	return "|\n" + indent + "  " + strings.ReplaceAll(command, "\n", "\n"+indent+"  ") + "\n"
}

func yamlValue(v string) string {
	if v == "true" || v == "false" {
		return v
	}
	return "'" + strings.ReplaceAll(v, "'", "''") + "'"
}

func missing(indent string, t protocol.Toolchain) string {
	return fmt.Sprintf("%s# smith: no plugin says how to install %s %s here; add a step in the before-setup slot\n", indent, t.Name, t.Version)
}
