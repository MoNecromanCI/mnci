package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/libs/plugin-jvm/discovery"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestGenerate(t *testing.T) {
	ws := t.TempDir()
	info := protocol.WorkspaceInfo{Name: "acme-corp"}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web-api", Directory: "apps/web-api", Workspace: info}); err != nil {
		t.Fatal(err)
	}
	res, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "core", Directory: "packages/core", Workspace: info, Stack: "java"})
	if err != nil || res.Files[len(res.Files)-1] != "pom.xml" {
		t.Fatal(err, res.Files)
	}
	root, _ := os.ReadFile(filepath.Join(ws, "pom.xml"))
	if !strings.Contains(string(root), "  <modules>\n    <module>apps/web-api</module>\n    <module>packages/core</module>\n  </modules>") || !strings.Contains(string(root), "<groupId>dev.acmecorp</groupId>") {
		t.Fatalf("root pom:\n%s", root)
	}
	if _, err := os.Stat(filepath.Join(ws, "apps/web-api/src/main/java/dev/acmecorp/webapi/Main.java")); err != nil {
		t.Fatal(err)
	}
	found, err := discovery.Discover(ws, []string{"pom.xml", "apps/web-api/pom.xml", "packages/core/pom.xml"})
	if err != nil || len(found.Projects) != 2 {
		t.Fatalf("%v %+v", err, found.Projects)
	}
	api, core := found.Projects[0], found.Projects[1]
	if api.Kind != protocol.ProjectApp || api.Exports[0] != "dev.acmecorp:acme-corp-web-api" || api.Tasks["start"].Command == nil {
		t.Errorf("api %+v", api)
	}
	if core.Kind != protocol.ProjectPackage || core.Stack != "java" || core.Version != "0.0.0" {
		t.Errorf("core %+v", core)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "packages/core", Workspace: info}); err == nil {
		t.Fatal("an existing directory is refused")
	}
}

func TestAddModuleWithoutList(t *testing.T) {
	out, ok := AddModule("<project>\n  <artifactId>x</artifactId>\n  <properties>\n  </properties>\n</project>\n", "libs/a")
	if !ok || !strings.Contains(out, "  <modules>\n    <module>libs/a</module>\n  </modules>\n\n  <properties>") {
		t.Fatalf("%s", out)
	}
	if _, ok := AddModule(out, "libs/a"); ok {
		t.Fatal("a listed module is not added twice")
	}
}
