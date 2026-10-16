package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestOutsideBuildDir(t *testing.T) {
	web := &projectgraph.Project{Plugin: "node", Project: protocol.Project{Name: "web", Root: "apps/web", Tasks: map[string]protocol.Task{
		"build": {Outputs: []string{"../../dist/apps/web"}},
		"test":  {Outputs: []string{"coverage"}},
	}}, Sources: map[string]string{"test.outputs": "smith.toml [tasks.test]"}}
	api := &projectgraph.Project{Plugin: "go", Project: protocol.Project{Name: "api", Root: "apps/api", Tasks: map[string]protocol.Task{"build": {Outputs: []string{"dist"}}}}}
	got := OutsideBuildDir([]*projectgraph.Project{web, api})
	want := []string{
		"web:test writes apps/web/coverage (from smith.toml [tasks.test]), not dist/apps/web/",
		"api:build writes apps/api/dist (from plugin go), not dist/apps/api/",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got:\n%s", strings.Join(got, "\n"))
	}
}
