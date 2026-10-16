package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestParseLists(t *testing.T) {
	l := ParseLists(`cmake_minimum_required(VERSION 3.25)
# project(ignored VERSION 9.9.9)
project(acme_api
  VERSION 1.2.0
  LANGUAGES CXX)
set(MNCI_KIND "app")
set(MNCI_STACK cpp)
add_subdirectory(${CMAKE_CURRENT_SOURCE_DIR}/../../libs/core ${CMAKE_CURRENT_BINARY_DIR}/core)
add_subdirectory(src)
add_executable(acme_api src/main.cpp)
`)
	if !l.Project || l.Name != "acme_api" || l.Version != "1.2.0" || l.Kind != "app" || l.Stack != "cpp" || l.Executable != "acme_api" || len(l.Subdirectories) != 2 {
		t.Fatalf("%+v", l)
	}
	if r := SubdirectoryRoot("apps/api", l.Subdirectories[0]); r != "libs/core" {
		t.Fatal(r)
	}
	if r := SubdirectoryRoot("apps/api", "../../../outside"); r != "" {
		t.Fatal(r)
	}
}

func TestDiscover(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("apps/api/CMakeLists.txt", "project(acme_api VERSION 0.1.0)\nset(MNCI_KIND app)\nadd_subdirectory(../../libs/core core)\nadd_subdirectory(src)\nadd_executable(acme_api src/main.cpp)\n")
	write("apps/api/src/CMakeLists.txt", "target_sources(acme_api PRIVATE x.cpp)\n")
	write("apps/api/src/main.cpp", "")
	write("libs/core/CMakeLists.txt", "project(acme_core)\nadd_library(acme_core src/core.cpp)\n")
	write("libs/core/src/core.cpp", "")
	write("libs/core/include/acme_core/core.hpp", "")
	res, err := Discover(ws, []string{"apps/api/CMakeLists.txt", "apps/api/src/CMakeLists.txt", "libs/core/CMakeLists.txt"})
	if err != nil || len(res.Projects) != 2 {
		t.Fatalf("a subdirectory without project() is not a project: %v %+v", err, res.Projects)
	}
	api, core := res.Projects[0], res.Projects[1]
	if api.Kind != protocol.ProjectApp || api.Version != "0.1.0" || len(api.Imports) != 1 || api.Imports[0] != "libs/core" || core.Exports[0] != "libs/core" {
		t.Errorf("api %+v core %+v", api, core)
	}
	if c := api.Tasks["configure"].Command; c[4] != "{workspaceRoot}/dist/apps/api" || c[6] != "Ninja" {
		t.Errorf("configure %q", c)
	}
	if api.Tasks["start"].DependsOn[0] != "build" || api.Tasks["test"].DependsOn[0] != "build" {
		t.Errorf("api tasks %+v", api.Tasks)
	}
	if lint := core.Tasks["lint"].Command; lint[len(lint)-2] != "src" || lint[len(lint)-1] != "include" {
		t.Errorf("core lint %q", lint)
	}
}
