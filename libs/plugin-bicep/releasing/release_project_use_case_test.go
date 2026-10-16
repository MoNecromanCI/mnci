package releasing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestSetVersion(t *testing.T) {
	if out, ok := SetVersion("metadata version = '0.0.0'\nparam x string\n", "1.0.0"); !ok || out != "metadata version = '1.0.0'\nparam x string\n" {
		t.Errorf("%q", out)
	}
	if out, ok := SetVersion("param x string\n", "1.0.0"); !ok || out != "metadata version = '1.0.0'\nparam x string\n" {
		t.Errorf("added at the top: %q", out)
	}
	if _, ok := SetVersion("metadata version = '1.0.0'\n", "1.0.0"); ok {
		t.Error("the same version is no change")
	}
}

func TestWriteVersion(t *testing.T) {
	ws := t.TempDir()
	file := filepath.Join(ws, "apps", "infra", "main.bicep")
	_ = os.MkdirAll(filepath.Dir(file), 0o755)
	_ = os.WriteFile(file, []byte("metadata version = '0.0.0'\n"), 0o644)
	project := protocol.Project{Root: "apps/infra"}
	if res, err := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "0.1.0", DryRun: true}); err != nil || res.Files[0] != "apps/infra/main.bicep" {
		t.Fatal(res, err)
	}
	if raw, _ := os.ReadFile(file); string(raw) != "metadata version = '0.0.0'\n" {
		t.Fatal("a dry run wrote the version")
	}
	if _, err := WriteVersion(ws, protocol.VersionParams{Project: project, Version: "0.1.0"}); err != nil {
		t.Fatal(err)
	}
	if raw, _ := os.ReadFile(file); string(raw) != "metadata version = '0.1.0'\n" {
		t.Fatalf("%q", raw)
	}
	if res, _ := Publish(protocol.PublishParams{}); res.Skipped == "" {
		t.Error("nothing is published")
	}
}
