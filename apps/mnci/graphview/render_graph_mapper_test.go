package graphview

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/apps/mnci/projectgraph"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func sample(t *testing.T) *projectgraph.Graph {
	t.Helper()
	mk := func(name string, kind protocol.ProjectKind, deps ...string) *projectgraph.Project {
		return &projectgraph.Project{Project: protocol.Project{Name: name, Root: "x/" + name, Kind: kind, Language: "node", Tasks: map[string]protocol.Task{"build": {}, "test": {}}}, Deps: deps, Plugin: "node"}
	}
	g, err := projectgraph.New([]*projectgraph.Project{mk("web", protocol.ProjectApp, "ui", "sdk"), mk("ui", protocol.ProjectLib), mk("sdk", protocol.ProjectPackage, "ui")})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestRenderers(t *testing.T) {
	g := sample(t)
	text := Text(g)
	if !strings.Contains(text, "web  app · node · x/web\n  ├─ ui\n  └─ sdk\n") && !strings.Contains(text, "  ├─ sdk\n  └─ ui\n") {
		t.Fatalf("text:\n%s", text)
	}
	var parsed struct {
		Projects []Node `json:"projects"`
	}
	if err := json.Unmarshal([]byte(JSON(g)), &parsed); err != nil || len(parsed.Projects) != 3 || parsed.Projects[0].Tasks[0] != "build" {
		t.Fatalf("json %v %+v", err, parsed)
	}
	if dot := DOT(g); !strings.Contains(dot, `"web" -> "ui"`) {
		t.Fatalf("dot:\n%s", dot)
	}
	page := HTML(g, "acme <&>")
	for _, want := range []string{"<svg", "acme &lt;&amp;&gt;", "prefers-color-scheme: dark", "marker-end"} {
		if !strings.Contains(page, want) {
			t.Errorf("html lacks %q", want)
		}
	}
	if strings.Contains(page, "acme <&>") {
		t.Fatal("html must escape the title")
	}
}
