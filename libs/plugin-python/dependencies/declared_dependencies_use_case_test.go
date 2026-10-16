package dependencies

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDeclaredAndRewrite(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "apps", "api"), 0o755)
	file := filepath.Join(ws, "apps", "api", "pyproject.toml")
	_ = os.WriteFile(file, []byte("[project]\nname = \"api\"\ndependencies = [\"Requests>=2.31\", \"httpx>=0.27,<1\", \"rich\"]\n\n[project.optional-dependencies]\ndev = [\"pytest==8.2.0\"]\n"), 0o644)
	res, err := Declared(ws, protocol.DeclaredParams{Projects: []protocol.Project{{Name: "api", Root: "apps/api"}}})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]protocol.Declaration{}
	for _, d := range res.Declarations {
		got[d.Name] = d
	}
	if d := got["requests"]; d.Spec != ">=2.31" || !d.Rewritable || d.Section != "dependencies" || d.File != "apps/api/pyproject.toml" {
		t.Errorf("requests: %+v", d)
	}
	if got["httpx"].Rewritable || got["rich"].Rewritable {
		t.Errorf("a compound range or no constraint is not rewritable: %+v %+v", got["httpx"], got["rich"])
	}
	if d := got["pytest"]; d.Section != "dev" || d.Spec != "==8.2.0" {
		t.Errorf("pytest: %+v", d)
	}
	_, err = Rewrite(ws, protocol.RewriteParams{Changes: []protocol.Rewrite{{Declaration: got["requests"], Spec: ">=2.32.3"}, {Declaration: got["pytest"], Spec: "==8.3.0"}}})
	if err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(file)
	if !strings.Contains(string(after), `["Requests>=2.32.3", "httpx>=0.27,<1", "rich"]`) || !strings.Contains(string(after), `dev = ["pytest==8.3.0"]`) {
		t.Fatalf("rewrite:\n%s", after)
	}
}
