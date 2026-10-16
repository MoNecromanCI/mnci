package releasing

import "testing"

func TestSetScriptVersion(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"group = \"dev.acme\"\nversion = \"0.0.0\"\n", "group = \"dev.acme\"\nversion = \"1.0.0\"\n"},
		{"group = 'dev.acme'\nversion = '0.0.0'\n", "group = 'dev.acme'\nversion = \"1.0.0\"\n"},
		{"plugins {}\n\ngroup = \"dev.acme\"\n\ndependencies {}\n", "plugins {}\n\ngroup = \"dev.acme\"\nversion = \"1.0.0\"\n\ndependencies {}\n"},
		{"plugins {}", "plugins {}\n\nversion = \"1.0.0\"\n"},
	} {
		if got, changed := SetScriptVersion(c.in, "1.0.0"); got != c.want || !changed {
			t.Errorf("SetScriptVersion(%q) = %q", c.in, got)
		}
	}
	if _, changed := SetScriptVersion("version = \"1.0.0\"\n", "1.0.0"); changed {
		t.Error("the same version is no change")
	}
	// A version inside a block is not the module's.
	if got, _ := SetScriptVersion("plugins {\n    kotlin(\"jvm\") version \"2.4.21\"\n}\nversion = \"0.0.0\"\n", "1.0.0"); got != "plugins {\n    kotlin(\"jvm\") version \"2.4.21\"\n}\nversion = \"1.0.0\"\n" {
		t.Errorf("only the top-level version: %q", got)
	}
}

func TestSetPropertyVersion(t *testing.T) {
	if got, changed := SetPropertyVersion("mnci.kind=package\nversion=0.0.0\n", "1.0.0"); got != "mnci.kind=package\nversion=1.0.0\n" || !changed {
		t.Errorf("%q", got)
	}
}
