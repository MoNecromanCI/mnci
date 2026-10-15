package pipeline

import "github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"

// CI returns the Node contribution to every pipeline, or nothing when the
// workspace has no Node project.
func CI(params protocol.CIParams, nodeVersion string) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	result := protocol.CIResult{
		Toolchains: []protocol.Toolchain{{Name: "node", Version: nodeVersion}},
		Setup:      []protocol.Step{{Name: "Install Node dependencies", Command: []string{"npm", "ci", "--no-audit", "--no-fund"}, Blocking: true}},
		Audit:      []protocol.Step{{Name: "npm audit (runtime dependencies)", Command: []string{"npm", "audit", "--omit=dev", "--audit-level=moderate"}, Blocking: true}},
	}
	for _, p := range params.Projects {
		if p.Kind == protocol.ProjectPackage {
			result.Credentials = []protocol.Credential{{Env: "NODE_AUTH_TOKEN", Description: "npm token that can publish the workspace's packages", Required: true}}
			break
		}
	}
	return result
}
