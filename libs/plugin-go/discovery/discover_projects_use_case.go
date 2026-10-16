package discovery

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/projectkind"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Discover describes the module behind each go.mod. A go.mod at the workspace
// root is a single-module repository, not a project of a monorepo, and is
// skipped.
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	for _, manifest := range manifests {
		root := path.Dir(manifest)
		if root == "." {
			continue
		}
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(root))
		src, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err != nil {
			return result, err
		}
		mod, err := ParseModule(string(src))
		if err != nil {
			return result, fmt.Errorf("%s: %w", manifest, err)
		}
		imports, err := ScanImports(dir)
		if err != nil {
			return result, err
		}
		name := path.Base(root)
		kind := projectkind.Resolve(mod.Kind, root, protocol.ProjectLib)
		project := protocol.Project{
			Name:     name,
			Root:     root,
			Language: "go",
			Kind:     kind,
			Exports:  []string{mod.Path},
			Imports:  imports,
			Tasks:    Tasks(name, kind, mod.Native),
			Stack:    mod.Stack,
		}
		if mod.Native {
			project.Tags = append(project.Tags, NativeTag)
		}
		if kind == protocol.ProjectPackage {
			// Go resolves a nested module's version from a <dir>/vX.Y.Z tag.
			project.TagPrefix = root + "/v"
		}
		result.Projects = append(result.Projects, project)
	}
	return result, nil
}
