package releasing

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/toolrun"
)

// npm is the command runner; tests replace it.
var npm = struct {
	output func(dir string, argv []string) (string, error)
	run    func(dir string, argv []string) error
}{
	output: func(dir string, argv []string) (string, error) { return toolrun.Output(dir, argv) },
	run:    func(dir string, argv []string) error { return toolrun.Run(dir, argv) },
}

// Publish publishes one package with npm. A version the registry already has
// is skipped, which is what makes a re-run after a partial release safe.
func Publish(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	if len(params.Project.Exports) == 0 {
		return protocol.PublishResult{}, errors.New("package.json has no name")
	}
	name := params.Project.Exports[0]
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root))
	registry, _ := npm.output(dir, []string{"npm", "config", "get", "registry"})
	result := protocol.PublishResult{Registry: strings.TrimSpace(registry)}
	if params.DryRun {
		result.Skipped = "dry run"
		return result, nil
	}
	existing, _ := npm.output(dir, []string{"npm", "view", name + "@" + params.Version, "version"})
	if strings.TrimSpace(existing) == params.Version {
		result.Skipped = "already published"
		return result, nil
	}
	if err := npm.run(dir, []string{"npm", "publish", "--access", "public"}); err != nil {
		return result, err
	}
	result.Published = true
	return result, nil
}
