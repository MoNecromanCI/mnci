package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCIAndDoctor(t *testing.T) {
	ws := t.TempDir()
	api := protocol.Project{Name: "api", Root: "apps/api", Kind: protocol.ProjectApp}
	if res := CI(ws, protocol.CIParams{Projects: []protocol.Project{api}}); len(res.Credentials) != 0 {
		t.Fatal("no Dockerfile, no registry")
	}
	_ = os.MkdirAll(filepath.Join(ws, "apps", "api"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "apps", "api", "Dockerfile"), nil, 0o644)
	if res := CI(ws, protocol.CIParams{Projects: []protocol.Project{api}}); len(res.Credentials) != 3 || res.Credentials[0].Env != "CONTAINER_REGISTRY" || res.Credentials[0].Required {
		t.Fatalf("no registry setting keeps CONTAINER_REGISTRY, optional: %+v", res)
	}
	api.Settings = map[string]map[string]any{"docker": {"registry": "ghcr.io/acme"}}
	if res := CI(ws, protocol.CIParams{Projects: []protocol.Project{api}}); len(res.Credentials) != 1 || res.Credentials[0].Env != "GITHUB_TOKEN" || len(res.Permissions) != 1 || res.Permissions[0] != "packages: write" {
		t.Fatalf("ghcr.io: %+v", res)
	}
	api.Settings["docker"]["push"] = false
	if res := CI(ws, protocol.CIParams{Projects: []protocol.Project{api}}); len(res.Credentials) != 0 || len(res.Permissions) != 0 {
		t.Fatalf("an image not pushed needs nothing: %+v", res)
	}
	real := probe
	defer func() { probe = real }()
	probe = func(string, []string) (string, error) { return "", errors.New("no daemon") }
	if res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{api}}); len(res.Findings) != 1 || res.Findings[0].OK {
		t.Fatalf("%+v", res)
	}
}
