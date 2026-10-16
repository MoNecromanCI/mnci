package images

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestContribute(t *testing.T) {
	ws := t.TempDir()
	for _, dir := range []string{"apps/API", "apps/cli", "libs/core"} {
		_ = os.MkdirAll(filepath.Join(ws, filepath.FromSlash(dir)), 0o755)
	}
	_ = os.WriteFile(filepath.Join(ws, "apps", "API", Dockerfile), nil, 0o644)
	_ = os.WriteFile(filepath.Join(ws, "libs", "core", Dockerfile), nil, 0o644)
	build := map[string]protocol.Task{"build": {}}
	res := Contribute(ws, protocol.ContributeParams{Projects: []protocol.Project{
		{Name: "API", Root: "apps/API", Kind: protocol.ProjectApp, Version: "1.2.0", Tasks: build},
		{Name: "cli", Root: "apps/cli", Kind: protocol.ProjectApp, Tasks: build},
		{Name: "core", Root: "libs/core", Kind: protocol.ProjectLib},
	}})
	if len(res.Projects) != 1 {
		t.Fatalf("only an app with a Dockerfile ships: %+v", res.Projects)
	}
	image := res.Projects["API"].Tasks[TaskName]
	if got := strings.Join(image.Command, " "); got != "docker build -f apps/API/Dockerfile -t api:1.2.0 -t api:latest ." {
		t.Errorf("command %q", got)
	}
	if image.Cwd != "." || image.DependsOn[0] != "build" || image.Cache {
		t.Errorf("image %+v", image)
	}
}
