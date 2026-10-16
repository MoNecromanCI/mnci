package releasing

import (
	"os"
	"path"
	"path/filepath"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// WriteVersion records a release's version in the project() call.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	rel := path.Join(params.Project.Root, "CMakeLists.txt")
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

// Publish has no registry to push to: a C++ project, a package too, is
// released by its tag (a vcpkg registry or Conan is the consumer's choice).
func Publish(protocol.PublishParams) (protocol.PublishResult, error) {
	return protocol.PublishResult{Skipped: "released by its tag"}, nil
}
