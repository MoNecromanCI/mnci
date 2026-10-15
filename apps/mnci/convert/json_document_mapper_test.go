package convert

import (
	"strings"
	"testing"
)

func TestDocumentKeepsOrderAndEditsInPlace(t *testing.T) {
	src := `{
  "name": "@acme/web",
  "version": "1.0.0",
  "nx": {
    "targets": {
      "build": { "executor": "x" }
    }
  },
  "files": ["dist", "!**/*.map"],
  "scripts": {
    "test": "jest",
    "e2e": "a && b <c>"
  },
  "count": 1.50
}
`
	doc, err := ParseDocument([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	doc.Delete("nx")
	doc.Object("scripts", false).SetString("build", "tsc -p tsconfig.lib.json")
	doc.Object("scripts", false).SetString("test", "jest --ci")
	doc.Object("mnci", true).SetString("kind", "package")
	want := `{
  "name": "@acme/web",
  "version": "1.0.0",
  "files": [
    "dist",
    "!**/*.map"
  ],
  "scripts": {
    "test": "jest --ci",
    "e2e": "a && b <c>",
    "build": "tsc -p tsconfig.lib.json"
  },
  "count": 1.50,
  "mnci": {
    "kind": "package"
  }
}
`
	if got := string(doc.Bytes()); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	if _, err := ParseDocument([]byte("[1]")); err == nil || !strings.Contains(err.Error(), "object") {
		t.Fatalf("an array is not a document: %v", err)
	}
	empty, _ := ParseDocument([]byte(`{"a": {}}`))
	if string(empty.Bytes()) != "{\n  \"a\": {}\n}\n" {
		t.Fatalf("%q", empty.Bytes())
	}
}
