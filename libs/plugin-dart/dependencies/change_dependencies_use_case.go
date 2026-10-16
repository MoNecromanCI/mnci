package dependencies

import (
	"path"
	"path/filepath"

	"github.com/MoNecromanCI/mnci/libs/plugin-dart/scaffolding"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// tool is flutter for a Flutter project (its tasks run flutter), else dart.
func tool(p protocol.Project) string {
	for _, t := range p.Tasks {
		if len(t.Command) > 0 && t.Command[0] == "flutter" {
			return "flutter"
		}
	}
	return "dart"
}

// Install resolves the whole workspace, or adds packages to one project
// with pub add (dev: prefixes them for dev_dependencies).
func Install(workspaceRoot string, params protocol.InstallParams) (protocol.DependencyResult, error) {
	if params.Project == nil || len(params.Packages) == 0 {
		return protocol.DependencyResult{}, scaffolding.Resolve(workspaceRoot)
	}
	p := *params.Project
	argv := []string{tool(p), "pub", "add"}
	for _, pkg := range params.Packages {
		if params.Dev {
			pkg = "dev:" + pkg
		}
		argv = append(argv, pkg)
	}
	if err := scaffolding.Runner(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root)), argv); err != nil {
		return protocol.DependencyResult{}, err
	}
	return protocol.DependencyResult{Files: []string{path.Join(p.Root, "pubspec.yaml"), "pubspec.lock"}}, nil
}

// Uninstall removes packages from one project with pub remove.
func Uninstall(workspaceRoot string, params protocol.UninstallParams) (protocol.DependencyResult, error) {
	argv := append([]string{tool(params.Project), "pub", "remove"}, params.Packages...)
	if err := scaffolding.Runner(filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root)), argv); err != nil {
		return protocol.DependencyResult{}, err
	}
	return protocol.DependencyResult{Files: []string{path.Join(params.Project.Root, "pubspec.yaml"), "pubspec.lock"}}, nil
}
