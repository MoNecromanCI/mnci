package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDoctorChecksThatAPackageDeclaresItsImports(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("package.json", `{"workspaces":["packages/*"]}`)
	write("packages/ocr/package.json", `{"name":"@acme/ocr"}`)
	write("packages/ocr/src/read.ts", `import type { Page } from '@acme/core'`)
	write("packages/web/package.json", `{"name":"@acme/web","private":true}`)
	write("packages/web/src/main.ts", `import { x } from 'undeclared'`)
	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{
		{Name: "ocr", Root: "packages/ocr", Kind: protocol.ProjectPackage},
		{Name: "web", Root: "packages/web", Kind: protocol.ProjectApp},
	}}, "24")
	var found *protocol.Finding
	for i, f := range res.Findings {
		if strings.HasPrefix(f.Check, "packages/web ") {
			t.Errorf("an app is not checked: %+v", f)
		}
		if f.Check == "packages/ocr declares the packages it imports" {
			found = &res.Findings[i]
		}
	}
	if found == nil || found.OK || found.Detail != "src/read.ts imports @acme/core, which package.json does not declare" {
		t.Fatalf("finding %+v", found)
	}
}
