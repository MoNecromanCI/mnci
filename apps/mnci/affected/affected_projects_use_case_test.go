package affected

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/MoNecromanCI/mnci/apps/mnci/gitclient"
	"github.com/MoNecromanCI/mnci/apps/mnci/projectgraph"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func setup(t *testing.T) (*gitclient.Git, *projectgraph.Graph) {
	t.Helper()
	g := gitclient.New(t.TempDir())
	if err := g.Init(); err != nil {
		t.Skip("git unavailable")
	}
	for _, kv := range [][2]string{{"user.email", "t@e.com"}, {"user.name", "T"}, {"commit.gpgsign", "false"}} {
		_, _ = g.Run("config", kv[0], kv[1])
	}
	write := func(rel string) {
		p := filepath.Join(g.Dir, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(rel+"x"), 0o644)
	}
	for _, f := range []string{"apps/web/a", "libs/ui/a", "libs/core/a", "apps/api/a", "README.md", "mnci.toml"} {
		write(f)
	}
	if err := g.CommitAll("init"); err != nil {
		t.Fatal(err)
	}
	_, _ = g.Run("branch", "base")
	mk := func(name, root string, deps ...string) *projectgraph.Project {
		return &projectgraph.Project{Project: protocol.Project{Name: name, Root: root}, Deps: deps}
	}
	pg, _ := projectgraph.New([]*projectgraph.Project{mk("web", "apps/web", "ui"), mk("ui", "libs/ui", "core"), mk("core", "libs/core"), mk("api", "apps/api")})
	return g, pg
}

func TestAffectedFollowsDependents(t *testing.T) {
	g, pg := setup(t)
	_ = os.WriteFile(filepath.Join(g.Dir, "libs", "ui", "a"), []byte("changed"), 0o644)
	_ = os.WriteFile(filepath.Join(g.Dir, "README.md"), []byte("docs only"), 0o644)
	r := Compute(g, pg, "base")
	if r.All || !reflect.DeepEqual(r.Projects, []string{"ui", "web"}) {
		t.Fatalf("result %+v", r)
	}
}

func TestAffectedFallsBackToEverything(t *testing.T) {
	g, pg := setup(t)
	if r := Compute(g, pg, "no-such-branch"); !r.All {
		t.Fatalf("an unknown base must verify everything: %+v", r)
	}
	if r := Compute(g, pg, ""); !r.All {
		t.Fatal("no base must verify everything")
	}
	_ = os.WriteFile(filepath.Join(g.Dir, "mnci.toml"), []byte("changed"), 0o644)
	if r := Compute(g, pg, "base"); !r.All || len(r.Projects) != 4 {
		t.Fatalf("a workspace file change must affect everything: %+v", r)
	}
}

func TestNoChanges(t *testing.T) {
	g, pg := setup(t)
	r := Compute(g, pg, "base")
	if r.All || len(r.Projects) != 0 {
		t.Fatalf("result %+v", r)
	}
}
