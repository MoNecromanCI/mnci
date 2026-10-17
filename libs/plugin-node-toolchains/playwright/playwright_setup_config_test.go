package playwright

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestSetup(t *testing.T) {
	s := setup("", protocol.GenerateParams{Directory: "apps/web"})
	if s.DevDependencies["@playwright/test"] != testVersion || s.Scripts[TaskName] != "playwright test" {
		t.Errorf("setup %+v", s)
	}
	cfg := s.Files["playwright.config.ts"]
	if strings.Contains(cfg, reportsToken) || !strings.Contains(cfg, "'../../dist/e2e/apps/web/report'") {
		t.Errorf("reports go to the e2e task's outputs, beside the build, not in it:\n%s", cfg)
	}
	// Vitest and Jest match *.test.* and *.spec.*: the starter must match neither.
	for name := range s.Files {
		if strings.Contains(name, ".test.") || strings.Contains(name, ".spec.") {
			t.Errorf("%s would also run as a unit test", name)
		}
	}
	if _, ok := s.Files["e2e/smoke.e2e.ts"]; !ok {
		t.Errorf("files %v", s.Files)
	}
}

func TestDescribe(t *testing.T) {
	res, err := New("1.0.0").Describe(protocol.InitializeParams{WorkspaceRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Provides) != 1 || res.Provides[0] != TaskName || len(res.Stacks) != 1 || res.Stacks[0].Name != Stack {
		t.Errorf("describe %+v", res)
	}
}
