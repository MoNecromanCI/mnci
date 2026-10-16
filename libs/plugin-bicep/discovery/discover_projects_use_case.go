package discovery

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Discover describes the project behind each main.bicep. A project exports
// its directory; the workspace directories its modules come from, outside
// it, are its imports. A deployment is an app unless its metadata says otherwise.
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	for _, manifest := range manifests {
		root := path.Dir(manifest)
		if root == "." {
			continue
		}
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(root))
		src, err := os.ReadFile(filepath.Join(dir, Manifest))
		if err != nil {
			return result, err
		}
		file := ParseBicep(string(src))
		var imports []string
		for _, m := range file.Modules {
			if r := ModuleRoot(root, m); r != "" && r != root && !strings.HasPrefix(r, root+"/") {
				imports = append(imports, r)
			}
		}
		_, err = os.Stat(filepath.Join(dir, "main.bicepparam"))
		result.Projects = append(result.Projects, protocol.Project{
			Name:     path.Base(root),
			Root:     root,
			Language: "bicep",
			Kind:     projectkind.Resolve(file.Kind, root, protocol.ProjectApp),
			Exports:  []string{root},
			Imports:  imports,
			Version:  file.Version,
			Tasks:    Tasks(root, err == nil),
			Stack:    file.Stack,
		})
	}
	return result, nil
}
