package workspace

import "testing"

func TestRemoveProjectTables(t *testing.T) {
	src := `[workspace]
name = "acme" # keep

[projects.web]
tag_prefix = "web-v"

[projects.web.tasks.build]
command = ["make"]

[projects.webkit]
stack = "node-ts"

[projects."web".release]
product = "web"

[tasks.lint]
cache = false
`
	want := `[workspace]
name = "acme" # keep

[projects.webkit]
stack = "node-ts"

[tasks.lint]
cache = false
`
	got, removed := RemoveProjectTables(src, "web")
	if !removed || got != want {
		t.Fatalf("removed=%v\n%s", removed, got)
	}
	if _, err := Parse([]byte(got)); err != nil {
		t.Fatal(err)
	}
	if _, removed := RemoveProjectTables(want, "api"); removed {
		t.Fatal("nothing to remove for api")
	}
}
