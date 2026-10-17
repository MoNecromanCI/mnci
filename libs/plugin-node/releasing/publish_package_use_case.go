package releasing

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// npm is the command runner; tests replace it.
var npm = struct {
	output func(dir string, argv []string) (string, error)
	run    func(dir string, argv []string) error
}{
	output: func(dir string, argv []string) (string, error) { return toolrun.Output(dir, argv) },
	run:    func(dir string, argv []string) error { return toolrun.Run(dir, argv) },
}

// Publish publishes one package with npm, or a VS Code extension to the
// Marketplace, or builds an app's release asset. A version the registry
// already has is skipped, which is what makes a re-run after a partial
// release safe.
func Publish(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	if params.Assets != nil {
		return BuildAssets(workspaceRoot, params)
	}
	// An app is deployed, never published to npm: it releases by its tag,
	// its assets and its deploy plugins (an image).
	if params.Project.Kind == protocol.ProjectApp {
		return protocol.PublishResult{Skipped: "released by its tag"}, nil
	}
	if len(params.Project.Exports) == 0 {
		return protocol.PublishResult{}, errors.New("package.json has no name")
	}
	name := params.Project.Exports[0]
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root))
	if _, ok := Extension(dir); ok {
		return publishExtension(workspaceRoot, params)
	}
	// npm refuses config and view inside a workspace package, so ask at the root.
	registry, _ := npm.output(workspaceRoot, []string{"npm", "config", "get", "registry"})
	result := protocol.PublishResult{Registry: strings.TrimSpace(registry)}
	if params.DryRun {
		result.Skipped = "dry run"
		return result, nil
	}
	existing, _ := npm.output(workspaceRoot, []string{"npm", "view", name + "@" + params.Version, "version"})
	if strings.TrimSpace(existing) == params.Version {
		result.Skipped = "already published"
		return result, nil
	}
	// A project that builds into dist/ is published from there, staged.
	staged, err := Stage(workspaceRoot, params.Project.Root)
	if err != nil {
		return result, err
	}
	if staged == dir {
		err = npm.run(dir, []string{"npm", "publish", "--access", "public"})
	} else {
		err = npm.run(workspaceRoot, []string{"npm", "publish", staged, "--access", "public"})
	}
	if err != nil {
		return result, err
	}
	result.Published = true
	return result, nil
}
