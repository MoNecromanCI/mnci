package releasing

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MoNecromanCI/mnci/libs/plugin-python/venv"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/toolrun"
)

// Twine runs twine and returns its combined output; tests replace it.
var Twine = func(dir string, argv []string, env []string) (string, error) {
	out, err := toolrun.Output(dir, argv, env...)
	_, _ = toolrun.Log.Write([]byte(out))
	return out, err
}

// WriteVersion records the version in the project's pyproject.toml.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	manifest := path.Join(params.Project.Root, "pyproject.toml")
	full := filepath.Join(workspaceRoot, filepath.FromSlash(manifest))
	src, err := os.ReadFile(full)
	if err != nil {
		return protocol.VersionResult{}, err
	}
	out, err := SetVersion(string(src), params.Version)
	if err != nil {
		return protocol.VersionResult{}, err
	}
	if !params.DryRun {
		if err := os.WriteFile(full, []byte(out), 0o644); err != nil {
			return protocol.VersionResult{}, err
		}
	}
	return protocol.VersionResult{Files: []string{manifest}}, nil
}

// Publish uploads dist/* with twine. TWINE_REPOSITORY_URL picks the index;
// PYPI_TOKEN, when set, is used as an API token.
func Publish(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root))
	registry := os.Getenv("TWINE_REPOSITORY_URL")
	if registry == "" {
		registry = "https://upload.pypi.org/legacy/"
	}
	result := protocol.PublishResult{Registry: registry}
	if params.DryRun {
		result.Skipped = "dry run"
		return result, nil
	}
	dists, _ := filepath.Glob(filepath.Join(dir, "dist", "*"))
	sort.Strings(dists)
	var files []string
	for _, d := range dists {
		if strings.Contains(filepath.Base(d), "-"+params.Version) {
			files = append(files, d)
		}
	}
	if len(files) == 0 {
		return result, errors.New("no distribution for version " + params.Version + " in " + params.Project.Root + "/dist; build it first")
	}
	var env []string
	if token := os.Getenv("PYPI_TOKEN"); token != "" && os.Getenv("TWINE_PASSWORD") == "" {
		env = append(env, "TWINE_USERNAME=__token__", "TWINE_PASSWORD="+token)
	}
	argv := append([]string{venv.Path(workspaceRoot), "-m", "twine", "upload", "--non-interactive", "--skip-existing", "--disable-progress-bar"}, files...)
	out, err := Twine(dir, argv, env)
	if err != nil {
		return result, err
	}
	if strings.Contains(out, "Skipping") && !strings.Contains(out, "Uploading") {
		result.Skipped = "already published"
		return result, nil
	}
	result.Published = true
	return result, nil
}
