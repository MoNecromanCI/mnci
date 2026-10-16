package releasing

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestWriteVersion(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "apps", "cli"), 0o755)
	manifest := filepath.Join(ws, "apps", "cli", "Cargo.toml")
	_ = os.WriteFile(manifest, []byte("[package]\nname = \"acme-cli\"\nversion = \"0.0.0\"\n\n[dependencies]\ncore = { path = \"../../packages/core\", version = \"0.1.0\", package = \"acme-core\" }\nserde = { version = \"1\" }\n\n[workspace.metadata]\nversion = \"keep\"\n"), 0o644)
	res, err := WriteVersion(ws, protocol.VersionParams{Project: protocol.Project{Root: "apps/cli"}, Version: "1.2.0", Dependencies: map[string]string{"acme-core": "0.2.0", "serde": "9"}})
	if err != nil || len(res.Files) != 1 {
		t.Fatal(err, res)
	}
	got, _ := os.ReadFile(manifest)
	for _, want := range []string{"version = \"1.2.0\"\n\n[dependencies]", `version = "0.2.0", package = "acme-core"`, `serde = { version = "1" }`, "version = \"keep\""} {
		if !strings.Contains(string(got), want) {
			t.Errorf("lacks %q:\n%s", want, got)
		}
	}
	_ = os.WriteFile(manifest, []byte("[package]\nname = \"x\"\nversion.workspace = true\n"), 0o644)
	if res, _ := WriteVersion(ws, protocol.VersionParams{Project: protocol.Project{Root: "apps/cli"}, Version: "1.2.0"}); len(res.Files) != 0 {
		t.Fatal("an inherited version is the workspace's")
	}
}

func TestPublish(t *testing.T) {
	var ran []string
	real := CargoRunner
	defer func() { CargoRunner = real }()
	CargoRunner = func(_ string, argv []string) error { ran = append(ran, strings.Join(argv, " ")); return nil }
	core := protocol.Project{Name: "core", Kind: protocol.ProjectPackage, Exports: []string{"acme-core"}}
	if res, _ := Publish("/ws", protocol.PublishParams{Project: core, Version: "0.2.0", DryRun: true}); res.Published || ran[0] != "cargo publish -p acme-core --allow-dirty --dry-run" {
		t.Fatalf("%+v %v", res, ran)
	}
	if res, _ := Publish("/ws", protocol.PublishParams{Project: protocol.Project{Kind: protocol.ProjectLib}}); res.Skipped == "" {
		t.Fatal("a library is released by its tag")
	}
}

func TestPublished(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.URL.Path == "/crates/acme-core" {
			_, _ = w.Write([]byte(`{"crate": {"max_version": "0.3.0-rc.1", "max_stable_version": "0.2.0"}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	CratesAPI = srv.URL
	core := protocol.Project{Kind: protocol.ProjectPackage, Exports: []string{"acme-core"}}
	if res, _ := Published(protocol.PublishedParams{Project: core}); res.Version != "0.2.0" || res.Unreachable {
		t.Fatalf("%+v", res)
	}
	core.Exports = []string{"acme-new"}
	if res, _ := Published(protocol.PublishedParams{Project: core}); res.Version != "" || res.Unreachable {
		t.Fatalf("not published yet: %+v", res)
	}
}
