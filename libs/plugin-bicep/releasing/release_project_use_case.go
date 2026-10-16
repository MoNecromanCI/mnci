package releasing

import (
	"os"
	"path"
	"path/filepath"

	"github.com/MoNecromanCI/mnci/libs/plugin-bicep/discovery"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// WriteVersion records a release's version in main.bicep's metadata.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	rel := path.Join(params.Project.Root, discovery.Manifest)
	full := filepath.Join(workspaceRoot, filepath.FromSlash(rel))
	src, err := os.ReadFile(full)
	if err != nil {
		return protocol.VersionResult{}, err
	}
	out, changed := SetVersion(string(src), params.Version)
	if !changed {
		return protocol.VersionResult{Files: []string{}}, nil
	}
	if !params.DryRun {
		if err := os.WriteFile(full, []byte(out), 0o644); err != nil {
			return protocol.VersionResult{}, err
		}
	}
	return protocol.VersionResult{Files: []string{rel}}, nil
}

// Publish has no registry to push to: a deployment's templates are built
// into dist/ and deployed by the pipeline that owns the environment.
func Publish(protocol.PublishParams) (protocol.PublishResult, error) {
	return protocol.PublishResult{Skipped: "released by its tag"}, nil
}
