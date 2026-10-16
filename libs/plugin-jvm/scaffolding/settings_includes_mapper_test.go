package scaffolding

import "testing"

func TestAddInclude(t *testing.T) {
	out, ok := AddInclude("rootProject.name = \"acme\"", ":libs:a")
	if !ok || out != "rootProject.name = \"acme\"\n\ninclude(\":libs:a\")\n" {
		t.Fatalf("%q", out)
	}
	out, ok = AddInclude(out+"\nplugins {}\n", ":libs:b")
	if !ok || out != "rootProject.name = \"acme\"\n\ninclude(\":libs:a\")\ninclude(\":libs:b\")\n\nplugins {}\n" {
		t.Fatalf("after the last include: %q", out)
	}
	if _, ok := AddInclude(out, ":libs:a"); ok {
		t.Fatal("an included project is not added twice")
	}
	if RootProjectName("rootProject.name = 'acme'\n") != "acme" || RootProjectName("") != "" {
		t.Fatal("rootProject.name")
	}
}
