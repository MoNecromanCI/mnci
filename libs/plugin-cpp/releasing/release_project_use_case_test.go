package releasing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestSetVersion(t *testing.T) {
	if out, ok := SetVersion("project(acme_core\n  VERSION 0.0.0\n  LANGUAGES CXX)\n", "1.2.0"); !ok || out != "project(acme_core\n  VERSION 1.2.0\n  LANGUAGES CXX)\n" {
		t.Fatalf("%q", out)
	}
	if out, ok := SetVersion("project(acme_core LANGUAGES CXX)\n", "1.2.0"); !ok || out != "project(acme_core VERSION 1.2.0 LANGUAGES CXX)\n" {
		t.Fatalf("added %q", out)
	}
}

func TestWriteVersion(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "libs", "core"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "libs", "core", "CMakeLists.txt"), []byte("cmake_minimum_required(VERSION 3.25)\nproject(acme_core VERSION 0.0.0)\n"), 0o644)
	res, err := WriteVersion(ws, protocol.VersionParams{Project: protocol.Project{Root: "libs/core"}, Version: "0.1.0"})
	if err != nil || len(res.Files) != 1 {
		t.Fatal(err, res)
	}
	if src, _ := os.ReadFile(filepath.Join(ws, "libs", "core", "CMakeLists.txt")); string(src) != "cmake_minimum_required(VERSION 3.25)\nproject(acme_core VERSION 0.1.0)\n" {
		t.Fatalf("cmake_minimum_required's VERSION is not the project's:\n%s", src)
	}
}
