package pipeline

import (
	"os"
	"path/filepath"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// GovulncheckVersion is pinned; bump it deliberately.
const GovulncheckVersion = "v1.9.0"

// DefaultGoVersion is used when go.work does not say.
const DefaultGoVersion = "1.24"

// CI returns the Go contribution: the toolchain at go.work's version and a
// non-blocking vulnerability scan of every module.
func CI(workspaceRoot string, params protocol.CIParams) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	version := DefaultGoVersion
	if src, err := os.ReadFile(filepath.Join(workspaceRoot, "go.work")); err == nil {
		if v := ParseGoWork(string(src)).GoVersion; v != "" {
			version = v
		}
	}
	return protocol.CIResult{
		Toolchains: []protocol.Toolchain{{Name: "go", Version: version}},
		Setup:      []protocol.Step{{Name: "Download Go modules", Command: []string{"go", "mod", "download"}, Blocking: true}},
		Audit:      []protocol.Step{{Name: "govulncheck", Command: []string{"go", "run", "golang.org/x/vuln/cmd/govulncheck@" + GovulncheckVersion, "./..."}, Blocking: false}},
	}
}
