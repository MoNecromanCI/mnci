package pipeline

import (
	"github.com/monosmith-app/monosmith/libs/plugin-docker/images"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// shipping are the projects that ship as images.
func shipping(workspaceRoot string, projects []protocol.Project) []protocol.Project {
	var out []protocol.Project
	for _, p := range projects {
		if images.Ships(workspaceRoot, p) {
			out = append(out, p)
		}
	}
	return out
}

// CI returns what pushing the apps' images on release needs: the
// credentials of the registry smith.toml names (GITHUB_TOKEN with packages:
// write for ghcr.io), or nothing when no image is pushed. An app whose
// registry smith.toml does not name keeps CONTAINER_REGISTRY's optional
// variables, so a pipeline that sets it still pushes. Docker is on GitHub's
// and Azure's hosted Linux images.
func CI(workspaceRoot string, params protocol.CIParams) protocol.CIResult {
	var res protocol.CIResult
	seen := map[string]bool{}
	for _, p := range shipping(workspaceRoot, params.Projects) {
		t := images.TargetOf(p, func(string) string { return "" })
		if !images.Configured(p) && !images.PushOff(p) {
			t = images.Target{Legacy: true}
		}
		creds, perms := images.Credentials(t)
		for _, c := range creds {
			if !seen[c.Env] {
				seen[c.Env] = true
				res.Credentials = append(res.Credentials, c)
			}
		}
		for _, perm := range perms {
			if !seen[perm] {
				seen[perm] = true
				res.Permissions = append(res.Permissions, perm)
			}
		}
	}
	return res
}

// probe runs a tool for its output; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Output(dir, argv) }

// Doctor checks that Docker answers when an app ships as an image.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	if len(shipping(workspaceRoot, params.Projects)) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	_, err := probe(workspaceRoot, []string{"docker", "version", "--format", "{{.Server.Version}}"})
	return protocol.DoctorResult{Findings: []protocol.Finding{
		{Check: "Docker is running", OK: err == nil, Detail: "docker version could not reach a daemon", Remedy: "install and start Docker (https://docs.docker.com/get-docker/)"},
	}}
}
