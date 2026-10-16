package publishing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestPublish(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "apps", "api"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "apps", "api", "Dockerfile"), nil, 0o644)
	real := Docker
	defer func() { Docker = real }()
	var ran []string
	var password string
	Docker = func(_ string, argv []string, stdin string) error {
		ran = append(ran, strings.Join(argv, " "))
		if stdin != "" {
			password = stdin
		}
		return nil
	}
	api := protocol.Project{Name: "api", Root: "apps/api", Kind: protocol.ProjectApp}
	t.Setenv("CONTAINER_REGISTRY", "")
	if _, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"}); err == nil || !strings.Contains(err.Error(), "CONTAINER_REGISTRY") {
		t.Fatalf("a push without a registry says which: %v", err)
	}
	t.Setenv("CONTAINER_REGISTRY", "ghcr.io/acme/")
	if res, _ := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0", DryRun: true}); res.Skipped != "dry run" || len(ran) != 0 {
		t.Fatalf("a dry run runs nothing: %+v %v", res, ran)
	}
	t.Setenv("CONTAINER_USERNAME", "bot")
	t.Setenv("CONTAINER_PASSWORD", "s3cret")
	res, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"})
	if err != nil || !res.Published || res.Registry != "ghcr.io/acme" {
		t.Fatalf("%+v %v", res, err)
	}
	want := []string{
		"docker login ghcr.io -u bot --password-stdin",
		"docker build -f apps/api/Dockerfile -t ghcr.io/acme/api:1.0.0 -t ghcr.io/acme/api:latest .",
		"docker push ghcr.io/acme/api:1.0.0",
		"docker push ghcr.io/acme/api:latest",
	}
	if strings.Join(ran, "\n") != strings.Join(want, "\n") || password != "s3cret" {
		t.Fatalf("ran:\n%s", strings.Join(ran, "\n"))
	}
	if res, _ := Publish(ws, protocol.PublishParams{Project: protocol.Project{Name: "web", Root: "apps/web", Kind: protocol.ProjectApp}}); res.Skipped != "no Dockerfile" {
		t.Errorf("%+v", res)
	}
}
