package dependencies

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGoCommands(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "apps", "api"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "apps", "api", "go.mod"), []byte("module x\n\ngo 1.26\n\ntool (\n\tgolang.org/x/tools/cmd/stringer\n)\n"), 0o644)
	var ran []string
	real := GoRunner
	defer func() { GoRunner = real }()
	GoRunner = func(dir string, argv []string) error {
		rel, _ := filepath.Rel(ws, dir)
		ran = append(ran, filepath.ToSlash(rel)+": "+strings.Join(argv, " "))
		return nil
	}
	api := protocol.Project{Name: "api", Root: "apps/api"}
	res, _ := Install(ws, protocol.InstallParams{Project: &api, Packages: []string{"github.com/pkg/errors@v0.9.1"}})
	_, _ = Install(ws, protocol.InstallParams{Project: &api, Packages: []string{"golang.org/x/tools/cmd/stringer"}, Dev: true})
	_, _ = Install(ws, protocol.InstallParams{Projects: []protocol.Project{api, {Root: "libs/core"}}})
	_, _ = Uninstall(ws, protocol.UninstallParams{Project: api, Packages: []string{"github.com/pkg/errors", "golang.org/x/tools/cmd/stringer"}})
	want := []string{
		"apps/api: go get github.com/pkg/errors@v0.9.1",
		"apps/api: go get -tool golang.org/x/tools/cmd/stringer",
		"apps/api: go mod download",
		"libs/core: go mod download",
		"apps/api: go get -tool golang.org/x/tools/cmd/stringer@none",
		"apps/api: go get github.com/pkg/errors@none",
	}
	if strings.Join(ran, "\n") != strings.Join(want, "\n") {
		t.Fatalf("ran:\n%s", strings.Join(ran, "\n"))
	}
	if strings.Join(res.Files, ",") != "apps/api/go.mod,apps/api/go.sum" || len(res.Notes) != 1 {
		t.Fatalf("result %+v", res)
	}
}

func TestIsTool(t *testing.T) {
	gomod := "module x\n\ntool golang.org/x/a\n\ntool (\n\tgolang.org/x/b // codegen\n)\n\nrequire golang.org/x/c v1.0.0\n"
	for path, want := range map[string]bool{"golang.org/x/a": true, "golang.org/x/b": true, "golang.org/x/c": false, "golang.org/x": false} {
		if isTool(gomod, path) != want {
			t.Errorf("%s: want %v", path, want)
		}
	}
}
