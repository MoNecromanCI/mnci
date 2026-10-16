package pipeline

import (
	"github.com/MoNecromanCI/mnci/libs/plugin-docker/images"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/toolrun"
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

// CI returns the registry an app's image is pushed to on release. Docker is
// on GitHub's and Azure's hosted Linux images.
func CI(workspaceRoot string, params protocol.CIParams) protocol.CIResult {
	if len(shipping(workspaceRoot, params.Projects)) == 0 {
		return protocol.CIResult{}
	}
	return protocol.CIResult{Credentials: []protocol.Credential{
		{Env: "CONTAINER_REGISTRY", Description: "Registry the apps' images are pushed to (ghcr.io/<owner>)", Required: true},
		{Env: "CONTAINER_USERNAME", Description: "Registry user (omit to use the runner's docker login)"},
		{Env: "CONTAINER_PASSWORD", Description: "Registry password or token"},
	}}
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
