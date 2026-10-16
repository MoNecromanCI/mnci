package dependencies

import (
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestNpmCommands(t *testing.T) {
	var ran []string
	real := NpmRunner
	defer func() { NpmRunner = real }()
	NpmRunner = func(_ string, argv []string) error {
		ran = append(ran, strings.Join(argv, " "))
		return nil
	}
	web := protocol.Project{Name: "web", Root: "apps/web"}
	res, _ := Install("/ws", protocol.InstallParams{Project: &web, Packages: []string{"left-pad", "zod@^3"}, Dev: true})
	_, _ = Install("/ws", protocol.InstallParams{})
	_, _ = Uninstall("/ws", protocol.UninstallParams{Project: web, Packages: []string{"left-pad"}})
	want := []string{
		"npm install left-pad zod@^3 --workspace apps/web --save-dev --no-audit --no-fund",
		"npm install --no-audit --no-fund",
		"npm uninstall left-pad --workspace apps/web --no-audit --no-fund",
	}
	if strings.Join(ran, "\n") != strings.Join(want, "\n") {
		t.Fatalf("ran:\n%s", strings.Join(ran, "\n"))
	}
	if strings.Join(res.Files, ",") != "apps/web/package.json,package-lock.json" {
		t.Fatalf("files %v", res.Files)
	}
}
