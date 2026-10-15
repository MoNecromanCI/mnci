package projectkind

import (
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestResolve(t *testing.T) {
	cases := []struct {
		declared, root string
		want           protocol.ProjectKind
	}{
		{"package", "apps/x", protocol.ProjectPackage},
		{"", "apps/web", protocol.ProjectApp},
		{"", "libs/ui", protocol.ProjectLib},
		{"", "packages/sdk", protocol.ProjectPackage},
		{"bogus", "tools/x", protocol.ProjectLib},
	}
	for _, c := range cases {
		if got := Resolve(c.declared, c.root, protocol.ProjectLib); got != c.want {
			t.Errorf("Resolve(%q,%q)=%s want %s", c.declared, c.root, got, c.want)
		}
	}
	if Directory(protocol.ProjectApp) != "apps" || Directory(protocol.ProjectLib) != "libs" || Directory(protocol.ProjectPackage) != "packages" {
		t.Fatal("Directory mismatch")
	}
}
