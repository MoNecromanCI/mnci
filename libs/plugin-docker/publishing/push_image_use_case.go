package publishing

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-docker/images"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Docker runs docker; tests replace it. stdin, when set, is the command's
// input (the registry password for docker login).
var Docker = func(dir string, argv []string, stdin string) error {
	if stdin == "" {
		return toolrun.Run(dir, argv)
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout, cmd.Stderr = toolrun.Log, toolrun.Log
	return cmd.Run()
}

// Publish builds a released app's image tagged with its version and latest
// under the registry [conventions.docker] names, and pushes both. With no
// registry, or push = false, the image is not pushed and the release goes
// on. It logs in first: to ghcr.io with GITHUB_TOKEN (as GITHUB_ACTOR),
// elsewhere with CONTAINER_USERNAME and CONTAINER_PASSWORD; without them
// the daemon's own login is used.
func Publish(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	if !images.Ships(workspaceRoot, params.Project) {
		return protocol.PublishResult{Skipped: "no Dockerfile"}, nil
	}
	t := images.TargetOf(params.Project, os.Getenv)
	result := protocol.PublishResult{Registry: t.Registry}
	switch {
	case t.Registry == "":
		result.Skipped = "image built, not pushed: set [conventions.docker] registry in smith.toml to push it on release"
		return result, nil
	case !t.Push:
		result.Skipped = "image not pushed: push = false in smith.toml"
		return result, nil
	}
	if t.Legacy {
		_, _ = fmt.Fprintf(toolrun.Log, "docker: %s is deprecated; set [conventions.docker] registry = %q in smith.toml\n", images.LegacyRegistryEnv, t.Registry)
	}
	image := t.Reference()
	tags := []string{image + ":" + params.Version, image + ":latest"}
	if params.DryRun {
		result.Skipped = "would push " + tags[0] + " and " + tags[1]
		return result, nil
	}
	if err := login(workspaceRoot, t); err != nil {
		return result, err
	}
	if err := Docker(workspaceRoot, images.BuildArgs(params.Project, tags...), ""); err != nil {
		return result, err
	}
	for _, tag := range tags {
		if err := Docker(workspaceRoot, []string{"docker", "push", tag}, ""); err != nil {
			return result, err
		}
	}
	result.Published = true
	return result, nil
}

// login logs docker in to the target's registry with the credentials it
// takes, or leaves the daemon's login when none are set.
func login(workspaceRoot string, t images.Target) error {
	host := images.Host(t.Registry)
	if user := os.Getenv("CONTAINER_USERNAME"); user != "" {
		return Docker(workspaceRoot, []string{"docker", "login", host, "-u", user, "--password-stdin"}, os.Getenv("CONTAINER_PASSWORD"))
	}
	if images.LoginFor(t.Registry) == images.LoginGitHub && os.Getenv("GITHUB_TOKEN") != "" {
		user := os.Getenv("GITHUB_ACTOR")
		if user == "" {
			user = "x-access-token"
		}
		return Docker(workspaceRoot, []string{"docker", "login", host, "-u", user, "--password-stdin"}, os.Getenv("GITHUB_TOKEN"))
	}
	return nil
}
