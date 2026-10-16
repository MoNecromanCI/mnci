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
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("package.json", `{"name": "acme", "devDependencies": {"typescript": "~5.4.0"}}`)
	write("apps/web/package.json", "{\n  \"name\": \"@acme/web\",\n  \"dependencies\": {\n    \"zod\": \"^3.22.0\",\n    \"@acme/ui\": \"*\",\n    \"lib\": \"github:me/lib\"\n  },\n  \"peerDependencies\": {\n    \"react\": \"^18.0.0\"\n  }\n}\n")
	write("node_modules/zod/package.json", `{"version": "3.23.8"}`)
	res, err := Declared(ws, protocol.DeclaredParams{Projects: []protocol.Project{{Name: "web", Root: "apps/web"}}})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]protocol.Declaration{}
	for _, d := range res.Declarations {
		got[d.Name] = d
	}
	if d := got["typescript"]; d.Project != "" || d.File != "package.json" || !d.Rewritable {
		t.Errorf("root: %+v", d)
	}
	if d := got["zod"]; d.Project != "web" || d.Spec != "^3.22.0" || !d.Rewritable || res.Installed["zod"] != "3.23.8" {
		t.Errorf("zod: %+v installed %q", d, res.Installed["zod"])
	}
	for _, name := range []string{"@acme/ui", "lib", "react"} {
		if got[name].Rewritable {
			t.Errorf("%s must not be rewritable: %+v", name, got[name])
		}
	}
	if !got["react"].Peer {
		t.Error("react is a peer")
	}
	if _, err := Rewrite(ws, protocol.RewriteParams{Changes: []protocol.Rewrite{{Declaration: got["zod"], Spec: "^3.23.8"}}}); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(ws, "apps", "web", "package.json"))
	if !strings.Contains(string(after), "\"zod\": \"^3.23.8\",\n    \"@acme/ui\": \"*\"") {
		t.Fatalf("rewrite:\n%s", after)
	}
}
