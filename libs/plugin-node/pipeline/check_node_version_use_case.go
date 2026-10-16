package pipeline

import (
	"strings"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/toolrun"
)

// probe runs a tool for its version; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Output(dir, argv) }

// CheckNodeVersion compares the installed Node.js major with the one CI
// installs. A difference warns: the workspace may build differently in CI.
func CheckNodeVersion(workspaceRoot, pin string) protocol.Finding {
	f := protocol.Finding{Check: "Node.js is " + pin + ", as in CI", Warning: true, Remedy: "install Node.js " + pin + " (https://nodejs.org)"}
	got, err := probe(workspaceRoot, []string{"node", "--version"})
	if err != nil {
		f.Warning, f.Detail = false, "node is not on PATH"
		return f
	}
	have := strings.TrimPrefix(strings.TrimSpace(got), "v")
	major, _, _ := strings.Cut(have, ".")
	f.OK = major == pin
	f.Detail = "found Node.js " + have
	return f
}
