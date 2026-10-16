package dependencies

import (
	"path"
	"path/filepath"

	"github.com/MoNecromanCI/mnci/libs/plugin-php/scaffolding"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

var update = []string{"composer", "update", "--no-interaction", "--no-progress"}

// Install installs the workspace's packages; given packages, it first
// requires them in one package's composer.json (require-dev when dev), then
// updates the workspace so its vendor/ has them.
func Install(workspaceRoot string, params protocol.InstallParams) (protocol.DependencyResult, error) {
	if params.Project == nil || len(params.Packages) == 0 {
		return protocol.DependencyResult{Files: []string{"composer.lock"}}, scaffolding.Runner(workspaceRoot, []string{"composer", "install", "--no-interaction", "--no-progress"})
	}
	p := *params.Project
	argv := append([]string{"composer", "require"}, params.Packages...)
	if params.Dev {
		argv = append(argv, "--dev")
	}
	argv = append(argv, "--no-update", "--no-interaction")
	if err := scaffolding.Runner(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root)), argv); err != nil {
		return protocol.DependencyResult{}, err
	}
	files := []string{path.Join(p.Root, "composer.json"), "composer.lock"}
	return protocol.DependencyResult{Files: files}, scaffolding.Runner(workspaceRoot, update)
}

// Uninstall removes packages from one package's composer.json and updates
// the workspace.
func Uninstall(workspaceRoot string, params protocol.UninstallParams) (protocol.DependencyResult, error) {
	argv := append(append([]string{"composer", "remove"}, params.Packages...), "--no-update", "--no-interaction")
	if err := scaffolding.Runner(filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root)), argv); err != nil {
		return protocol.DependencyResult{}, err
	}
	files := []string{path.Join(params.Project.Root, "composer.json"), "composer.lock"}
	return protocol.DependencyResult{Files: files}, scaffolding.Runner(workspaceRoot, update)
}
