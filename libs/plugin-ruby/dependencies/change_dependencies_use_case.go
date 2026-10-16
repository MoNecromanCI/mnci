package dependencies

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-ruby/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func gemspecOf(workspaceRoot string, p protocol.Project) (string, error) {
	found, _ := filepath.Glob(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "*.gemspec"))
	if len(found) == 0 {
		return "", fmt.Errorf("%s has no gemspec", p.Root)
	}
	return path.Join(p.Root, filepath.Base(found[0])), nil
}

func edit(workspaceRoot string, p protocol.Project, change func(string) string) (string, error) {
	rel, err := gemspecOf(workspaceRoot, p)
	if err != nil {
		return "", err
	}
	full := filepath.Join(workspaceRoot, filepath.FromSlash(rel))
	src, err := os.ReadFile(full)
	if err != nil {
		return "", err
	}
	return rel, os.WriteFile(full, []byte(change(string(src))), 0o644)
}

// Install installs the workspace's bundle, after declaring packages in one
// gem's gemspec when it is given some.
func Install(workspaceRoot string, params protocol.InstallParams) (protocol.DependencyResult, error) {
	var res protocol.DependencyResult
	if params.Project != nil && len(params.Packages) > 0 {
		rel, err := edit(workspaceRoot, *params.Project, func(src string) string { return AddDependencies(src, params.Packages, params.Dev) })
		if err != nil {
			return res, err
		}
		res.Files = append(res.Files, rel)
	}
	if err := scaffolding.Runner(workspaceRoot, []string{"bundle", "install"}); err != nil {
		return res, err
	}
	res.Files = append(res.Files, "Gemfile.lock")
	return res, nil
}

// Uninstall takes packages out of one gem's gemspec and reinstalls the
// bundle.
func Uninstall(workspaceRoot string, params protocol.UninstallParams) (protocol.DependencyResult, error) {
	rel, err := edit(workspaceRoot, params.Project, func(src string) string { return RemoveDependencies(src, params.Packages) })
	if err != nil {
		return protocol.DependencyResult{}, err
	}
	if err := scaffolding.Runner(workspaceRoot, []string{"bundle", "install"}); err != nil {
		return protocol.DependencyResult{Files: []string{rel}}, err
	}
	return protocol.DependencyResult{Files: []string{rel, "Gemfile.lock"}}, nil
}
