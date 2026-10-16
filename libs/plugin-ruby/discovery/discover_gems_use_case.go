package discovery

import (
	"os"
	"path"
	"path/filepath"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/projectkind"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Discover describes the gem behind each gemspec (one per directory).
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	seen := map[string]bool{}
	for _, manifest := range manifests {
		root := path.Dir(manifest)
		if root == "." || seen[root] {
			continue
		}
		seen[root] = true
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(root))
		src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(manifest)))
		if err != nil {
			return result, err
		}
		gem := ParseGemspec(string(src))
		if gem.Name == "" {
			gem.Name = path.Base(root)
		}
		kind := projectkind.Resolve(gem.Kind, root, protocol.ProjectLib)
		executable := ""
		if len(gem.Executables) > 0 {
			executable = gem.Executables[0]
		}
		_, specErr := os.Stat(filepath.Join(dir, "spec"))
		result.Projects = append(result.Projects, protocol.Project{
			Name:     path.Base(root),
			Root:     root,
			Language: "ruby",
			Kind:     kind,
			Exports:  []string{gem.Name},
			Imports:  gem.Dependencies,
			Version:  gem.Version,
			Tasks:    Tasks(kind, executable, specErr == nil),
			Stack:    gem.Stack,
		})
	}
	return result, nil
}
