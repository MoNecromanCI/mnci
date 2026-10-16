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

func TestAddRootPlugin(t *testing.T) {
	line := `id("com.android.application") version "9.4.1" apply false`
	out, ok := AddRootPlugin("plugins {\n    kotlin(\"jvm\") version \"2.4.21\" apply false\n}\n", line)
	if !ok || out != "plugins {\n    kotlin(\"jvm\") version \"2.4.21\" apply false\n    "+line+"\n}\n" {
		t.Fatalf("%q", out)
	}
	if _, ok := AddRootPlugin(out, `id("com.android.application") version "9.5.0" apply false`); ok {
		t.Error("a pinned plugin keeps its version")
	}
	if out, ok := AddRootPlugin("", line); !ok || out != "\nplugins {\n    "+line+"\n}\n" {
		t.Errorf("a script without plugins gets a block: %q", out)
	}
}
