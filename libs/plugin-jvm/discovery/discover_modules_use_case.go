package discovery

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/projectkind"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Discover describes the module behind each pom.xml. The root pom (the
// aggregator and parent) and any other pom-packaged module only organise
// the build, so they are not projects.
func Discover(workspaceRoot string, manifests []string) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	for _, manifest := range manifests {
		root := path.Dir(manifest)
		if root == "." {
			continue
		}
		src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(manifest)))
		if err != nil {
			return result, err
		}
		pom, err := ParsePom(src)
		if err != nil {
			return result, fmt.Errorf("%s: %w", manifest, err)
		}
		if pom.Packaging == "pom" {
			continue
		}
		kind := projectkind.Resolve(pom.Kind, root, protocol.ProjectLib)
		result.Projects = append(result.Projects, protocol.Project{
			Name:     path.Base(root),
			Root:     root,
			Language: "jvm",
			Kind:     kind,
			Exports:  []string{pom.Coordinates()},
			Imports:  pom.Dependencies,
			Version:  pom.Version,
			Tasks:    Tasks(root, pom.Coordinates(), pom.ArtifactID, kind),
			Stack:    pom.Stack,
		})
	}
	return result, nil
}
