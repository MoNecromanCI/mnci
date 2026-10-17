package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDroppedDependenciesAreNamed(t *testing.T) {
	p := &projectgraph.Project{
		Project: protocol.Project{Name: "core"},
		Sources: map[string]string{"e2e.dependsOn": "smith.toml [projects.core.tasks.e2e]"},
		Dropped: map[string][]string{"e2e": {"build (from plugin node)"}},
	}
	got := droppedDependencies(p)
	if len(got) != 1 || !got[0].Warning || got[0].OK {
		t.Fatalf("findings %+v: one warning per task", got)
	}
	for _, want := range []string{"[projects.core.tasks.e2e]", "build (from plugin node)"} {
		if !strings.Contains(got[0].Detail, want) {
			t.Errorf("detail %q lacks %q", got[0].Detail, want)
		}
	}
	if !strings.Contains(got[0].Remedy, "smith explain core:e2e") {
		t.Errorf("remedy %q", got[0].Remedy)
	}
	if len(droppedDependencies(&projectgraph.Project{})) != 0 {
		t.Error("a project with nothing dropped has no finding")
	}
}
