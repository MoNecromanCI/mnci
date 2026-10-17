package publishing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// recordDocker replaces Docker for the test, recording each command and the
// password given on stdin.
func recordDocker(t *testing.T) (ran *[]string, password *string) {
	t.Helper()
	real := Docker
	t.Cleanup(func() { Docker = real })
	ran, password = &[]string{}, new(string)
	Docker = func(_ string, argv []string, stdin string) error {
		*ran = append(*ran, strings.Join(argv, " "))
		if stdin != "" {
			*password = stdin
		}
		return nil
	}
	return ran, password
}

func shipping(t *testing.T, settings map[string]any) (string, protocol.Project) {
	t.Helper()
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "apps", "api"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "apps", "api", "Dockerfile"), nil, 0o644)
	api := protocol.Project{Name: "api", Root: "apps/api", Kind: protocol.ProjectApp}
	if settings != nil {
		api.Settings = map[string]map[string]any{"docker": settings}
	}
	for _, env := range []string{"CONTAINER_REGISTRY", "CONTAINER_USERNAME", "CONTAINER_PASSWORD", "GITHUB_TOKEN", "GITHUB_ACTOR"} {
		t.Setenv(env, "")
	}
	return ws, api
}

func TestPublishWithoutARegistryBuildsButDoesNotPush(t *testing.T) {
	ws, api := shipping(t, nil)
	ran, _ := recordDocker(t)
	res, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"})
	if err != nil || res.Published || !strings.Contains(res.Skipped, "[conventions.docker] registry") || len(*ran) != 0 {
		t.Fatalf("%+v %v %v", res, err, *ran)
	}
	_, off := shipping(t, map[string]any{"registry": "ghcr.io/acme", "push": false})
	if res, err := Publish(ws, protocol.PublishParams{Project: off, Version: "1.0.0"}); err != nil || !strings.Contains(res.Skipped, "push = false") {
		t.Fatalf("push = false: %+v %v", res, err)
	}
}

func TestPublishToGHCRLogsInWithTheJobsToken(t *testing.T) {
	ws, api := shipping(t, map[string]any{"registry": "ghcr.io/acme/", "image": "acme-api"})
	ran, password := recordDocker(t)
	if res, _ := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0", DryRun: true}); res.Skipped != "would push ghcr.io/acme/acme-api:1.0.0 and ghcr.io/acme/acme-api:latest" || len(*ran) != 0 {
		t.Fatalf("a dry run names the image and runs nothing: %+v %v", res, *ran)
	}
	t.Setenv("GITHUB_TOKEN", "ghs_token")
	t.Setenv("GITHUB_ACTOR", "octocat")
	res, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"})
	if err != nil || !res.Published || res.Registry != "ghcr.io/acme" {
		t.Fatalf("%+v %v", res, err)
	}
	want := []string{
		"docker login ghcr.io -u octocat --password-stdin",
		"docker build -f apps/api/Dockerfile -t ghcr.io/acme/acme-api:1.0.0 -t ghcr.io/acme/acme-api:latest .",
		"docker push ghcr.io/acme/acme-api:1.0.0",
		"docker push ghcr.io/acme/acme-api:latest",
	}
	if strings.Join(*ran, "\n") != strings.Join(want, "\n") || *password != "ghs_token" {
		t.Fatalf("ran:\n%s", strings.Join(*ran, "\n"))
	}
}

func TestPublishToAnotherRegistryLogsInWithContainerCredentials(t *testing.T) {
	ws, api := shipping(t, map[string]any{"registry": "registry.example.com/team"})
	ran, password := recordDocker(t)
	t.Setenv("CONTAINER_USERNAME", "bot")
	t.Setenv("CONTAINER_PASSWORD", "s3cret")
	if _, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"}); err != nil {
		t.Fatal(err)
	}
	if (*ran)[0] != "docker login registry.example.com -u bot --password-stdin" || *password != "s3cret" {
		t.Fatalf("ran:\n%s", strings.Join(*ran, "\n"))
	}
}

func TestPublishStillReadsContainerRegistry(t *testing.T) {
	ws, api := shipping(t, nil)
	ran, _ := recordDocker(t)
	t.Setenv("CONTAINER_REGISTRY", "127.0.0.1:5000/e2e/")
	res, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"})
	if err != nil || !res.Published || res.Registry != "127.0.0.1:5000/e2e" || (*ran)[len(*ran)-1] != "docker push 127.0.0.1:5000/e2e/api:latest" {
		t.Fatalf("%+v %v %v", res, err, *ran)
	}
	if res, _ := Publish(ws, protocol.PublishParams{Project: protocol.Project{Name: "web", Root: "apps/web", Kind: protocol.ProjectApp}}); res.Skipped != "no Dockerfile" {
		t.Errorf("%+v", res)
	}
}
