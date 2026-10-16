package publishing

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/MoNecromanCI/mnci/libs/plugin-docker/images"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/toolrun"
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
// under CONTAINER_REGISTRY (ghcr.io/acme, 127.0.0.1:5000), and pushes both.
// CONTAINER_USERNAME and CONTAINER_PASSWORD log in first when set; without
// them the daemon's own login is used.
func Publish(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	if !images.Ships(workspaceRoot, params.Project) {
		return protocol.PublishResult{Skipped: "no Dockerfile"}, nil
	}
	registry := strings.TrimSuffix(os.Getenv("CONTAINER_REGISTRY"), "/")
	result := protocol.PublishResult{Registry: registry}
	if params.DryRun {
		result.Skipped = "dry run"
		return result, nil
	}
	if registry == "" {
		return result, errors.New("set CONTAINER_REGISTRY to the registry " + params.Project.Name + "'s image is pushed to")
	}
	if user := os.Getenv("CONTAINER_USERNAME"); user != "" {
		host, _, _ := strings.Cut(registry, "/")
		if err := Docker(workspaceRoot, []string{"docker", "login", host, "-u", user, "--password-stdin"}, os.Getenv("CONTAINER_PASSWORD")); err != nil {
			return result, err
		}
	}
	image := registry + "/" + images.Name(params.Project)
	tags := []string{image + ":" + params.Version, image + ":latest"}
	if err := Docker(workspaceRoot, images.BuildArgs(params.Project, tags...), ""); err != nil {
		return result, err
	}
	for _, t := range tags {
		if err := Docker(workspaceRoot, []string{"docker", "push", t}, ""); err != nil {
			return result, err
		}
	}
	result.Published = true
	return result, nil
}
