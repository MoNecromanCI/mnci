package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestDiscover(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("pom.xml", `<project><groupId>dev.acme</groupId><artifactId>acme-parent</artifactId><version>0.0.0</version><packaging>pom</packaging><modules><module>apps/api</module><module>libs/core</module></modules></project>`)
	write("apps/api/pom.xml", `<project>
  <parent><groupId>dev.acme</groupId><artifactId>acme-parent</artifactId><version>0.0.0</version></parent>
  <artifactId>acme-api</artifactId><version>1.2.0</version>
  <properties><mnci.kind>app</mnci.kind><exec.mainClass>dev.acme.api.Main</exec.mainClass></properties>
  <dependencies><dependency><groupId>${project.groupId}</groupId><artifactId>acme-core</artifactId></dependency><dependency><groupId>org.junit.jupiter</groupId><artifactId>junit-jupiter</artifactId></dependency></dependencies>
</project>`)
	write("libs/core/pom.xml", `<project><parent><groupId>dev.acme</groupId><version>0.0.0</version></parent><artifactId>acme-core</artifactId><properties><mnci.stack>java</mnci.stack></properties></project>`)
	res, err := Discover(ws, []string{"pom.xml", "apps/api/pom.xml", "libs/core/pom.xml"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 2 {
		t.Fatalf("the root pom is not a project: %+v", res.Projects)
	}
	api, core := res.Projects[0], res.Projects[1]
	if api.Kind != protocol.ProjectApp || api.Version != "1.2.0" || api.Exports[0] != "dev.acme:acme-api" || api.Imports[0] != "dev.acme:acme-core" || api.Tasks["start"].Command[5] != "exec:java" {
		t.Errorf("api %+v", api)
	}
	if build := api.Tasks["build"]; build.Cwd != "." || build.Command[4] != "dev.acme:acme-api" || build.Outputs[0] != "../../dist/maven/acme-api" || build.DependsOn[0] != "^build" {
		t.Errorf("api build %+v", build)
	}
	if core.Kind != protocol.ProjectLib || core.Exports[0] != "dev.acme:acme-core" || core.Version != "0.0.0" || core.Stack != "java" {
		t.Errorf("core %+v", core)
	}
}
