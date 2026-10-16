package dependencies

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/libs/plugin-php/scaffolding"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestInstall(t *testing.T) {
	ws := t.TempDir()
	var ran []string
	real := scaffolding.Runner
	defer func() { scaffolding.Runner = real }()
	scaffolding.Runner = func(dir string, argv []string) error {
		rel, _ := filepath.Rel(ws, dir)
		ran = append(ran, filepath.ToSlash(rel)+": "+strings.Join(argv, " "))
		return nil
	}
	if _, err := Install(ws, protocol.InstallParams{}); err != nil {
		t.Fatal(err)
	}
	cli := protocol.Project{Root: "apps/cli"}
	if _, err := Install(ws, protocol.InstallParams{Project: &cli, Packages: []string{"mockery/mockery"}, Dev: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := Uninstall(ws, protocol.UninstallParams{Project: cli, Packages: []string{"mockery/mockery"}}); err != nil {
		t.Fatal(err)
	}
	want := ".: composer install --no-interaction --no-progress|" +
		"apps/cli: composer require mockery/mockery --dev --no-update --no-interaction|.: composer update --no-interaction --no-progress|" +
		"apps/cli: composer remove mockery/mockery --no-update --no-interaction|.: composer update --no-interaction --no-progress"
	if strings.Join(ran, "|") != want {
		t.Fatalf("%q", ran)
	}
}
