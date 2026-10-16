package pipeline

import (
	"strings"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// runtime installs Go in the generated pipelines. The Go module cache is not
// cached by the action: mnci's own cache covers the tasks.
func runtime(version string) protocol.Toolchain {
	azure := version
	if strings.Count(azure, ".") == 1 {
		azure += ".0" // GoTool wants a full version
	}
	return protocol.Toolchain{
		Name: "go", Version: version,
		GitHub: &protocol.ActionStep{Uses: "actions/setup-go@v5", Inputs: []protocol.Input{{Name: "go-version", Value: version}, {Name: "cache", Value: "false"}}},
		Azure:  &protocol.ActionStep{Uses: "GoTool@0", Inputs: []protocol.Input{{Name: "version", Value: azure}}},
		// The feature would otherwise install the newest golangci-lint.
		Devcontainer: &protocol.DevcontainerFeature{
			ID:      "ghcr.io/devcontainers/features/go:1",
			Options: map[string]string{"version": azure, "golangciLintVersion": strings.TrimPrefix(GolangciLintVersion, "v")},
		},
	}
}
