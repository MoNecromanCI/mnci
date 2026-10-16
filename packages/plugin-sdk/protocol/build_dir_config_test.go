package protocol

import "testing"

func TestBuildDir(t *testing.T) {
	if ProjectBuildDir("apps/web") != "dist/apps/web" || BuildDirFromProject("apps/web") != "../../dist/apps/web" || BuildDirFromProject("tools/x/y") != "../../../dist/tools/x/y" {
		t.Fatal(ProjectBuildDir("apps/web"), BuildDirFromProject("apps/web"))
	}
	for p, want := range map[string]bool{"dist": true, "dist/apps/web": true, "distro/x": false, "apps/web/dist": false} {
		if InBuildDir(p) != want {
			t.Errorf("%s: %v", p, !want)
		}
	}
}
