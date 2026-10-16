package dependencies

import (
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/libs/plugin-dart/scaffolding"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestInstall(t *testing.T) {
	var ran []string
	real := scaffolding.Runner
	defer func() { scaffolding.Runner = real }()
	scaffolding.Runner = func(dir string, argv []string) error { ran = append(ran, strings.Join(argv, " ")); return nil }
	ws := t.TempDir()
	if _, err := Install(ws, protocol.InstallParams{}); err != nil {
		t.Fatal(err)
	}
	web := protocol.Project{Root: "apps/web", Tasks: map[string]protocol.Task{"lint": {Command: []string{"flutter", "analyze"}}}}
	res, err := Install(ws, protocol.InstallParams{Project: &web, Packages: []string{"mocktail"}, Dev: true})
	if err != nil || res.Files[0] != "apps/web/pubspec.yaml" {
		t.Fatal(err, res)
	}
	if _, err := Uninstall(ws, protocol.UninstallParams{Project: web, Packages: []string{"http"}}); err != nil {
		t.Fatal(err)
	}
	if strings.Join(ran, "|") != "dart pub get|flutter pub add dev:mocktail|flutter pub remove http" {
		t.Fatalf("%q", ran)
	}
}
