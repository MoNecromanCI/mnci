package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/libs/plugin-cpp/discovery"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestGenerate(t *testing.T) {
	ws := t.TempDir()
	info := protocol.WorkspaceInfo{Name: "acme-corp"}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core", Workspace: info, Stack: "cpp"}); err != nil {
		t.Fatal(err)
	}
	res, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web-api", Directory: "apps/web-api", Workspace: info})
	if err != nil || !strings.Contains(res.Notes[0], "add_subdirectory(${CMAKE_CURRENT_SOURCE_DIR}/../../apps/web-api acme_corp_web_api)") {
		t.Fatal(err, res)
	}
	lists, _ := os.ReadFile(filepath.Join(ws, "apps/web-api/CMakeLists.txt"))
	if !strings.Contains(string(lists), "add_executable(acme_corp_web_api src/main.cpp)\ntarget_link_libraries(acme_corp_web_api PRIVATE acme_corp_web_api_lib)") {
		t.Fatalf("CMakeLists.txt:\n%s", lists)
	}
	found, err := discovery.Discover(ws, []string{"libs/core/CMakeLists.txt", "apps/web-api/CMakeLists.txt"})
	if err != nil || len(found.Projects) != 2 {
		t.Fatalf("%v %+v", err, found.Projects)
	}
	core, api := found.Projects[0], found.Projects[1]
	if core.Kind != protocol.ProjectLib || core.Stack != "cpp" || core.Version != "0.0.0" || core.Tasks["lint"].Command == nil {
		t.Errorf("core %+v", core)
	}
	if api.Kind != protocol.ProjectApp || !strings.HasSuffix(strings.TrimSuffix(api.Tasks["start"].Command[0], ".exe"), "/dist/apps/web-api/acme_corp_web_api") {
		t.Errorf("api %+v", api)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "libs/core", Workspace: info}); err == nil {
		t.Fatal("an existing directory is refused")
	}
}
