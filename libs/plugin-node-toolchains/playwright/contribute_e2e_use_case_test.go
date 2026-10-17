package playwright

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func writeProject(t *testing.T, ws, dir, manifest string, files ...string) {
	t.Helper()
	root := filepath.Join(ws, filepath.FromSlash(dir))
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(root, f), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func onPlatform(t *testing.T, goos string, ci bool) {
	t.Helper()
	realPlatform, realCI := platform, inCI
	platform, inCI = goos, func() bool { return ci }
	t.Cleanup(func() { platform, inCI = realPlatform, realCI })
}

var e2eScript = protocol.Task{Command: []string{"npm", "run", "e2e", "--silent"}}

func TestContributeDetectsEachFlavour(t *testing.T) {
	onPlatform(t, "darwin", false)
	ws := t.TempDir()
	writeProject(t, ws, "apps/web", `{"devDependencies": {"@playwright/test": "^1.63.0"}}`)
	writeProject(t, ws, "apps/site", `{}`, "playwright.config.ts")
	writeProject(t, ws, "packages/core", `{"dependencies": {"playwright": "^1.63.0"}}`)
	writeProject(t, ws, "packages/office", `{"dependencies": {"playwright-core": "^1.63.0"}}`)
	writeProject(t, ws, "packages/plain", `{"dependencies": {"zod": "^4.0.0"}}`)
	res := Contribute(ws, protocol.ContributeParams{Projects: []protocol.Project{
		{Name: "web", Root: "apps/web", Tasks: map[string]protocol.Task{"build": {}}},
		{Name: "site", Root: "apps/site"},
		{Name: "core", Root: "packages/core", Tasks: map[string]protocol.Task{TaskName: e2eScript, "build": {}}},
		// The library alone says nothing about how its own runner starts: no e2e script, no task.
		{Name: "office", Root: "packages/office"},
		{Name: "plain", Root: "packages/plain", Tasks: map[string]protocol.Task{TaskName: e2eScript}},
	}})
	if len(res.Projects) != 3 {
		t.Fatalf("web, site and core have suites: %+v", res.Projects)
	}
	web := res.Projects["web"].Replace[TaskName]
	if got := strings.Join(web.Command, " "); got != "npx --no-install playwright test" {
		t.Errorf("a runner suite with no e2e script runs the runner: %q", got)
	}
	if got := strings.Join(web.DependsOn, " "); got != BrowsersTask+" ^build build" {
		t.Errorf("web e2e depends on %q", got)
	}
	if !web.Cache || !strings.Contains(web.Description, "runner, chromium") {
		t.Errorf("web e2e %+v", web)
	}
	if len(web.Outputs) != 1 || web.Outputs[0] != "../../dist/e2e/apps/web" {
		t.Errorf("the reports are the task's outputs, outside the build directory: %v", web.Outputs)
	}
	if got := strings.Join(res.Projects["site"].Replace[TaskName].DependsOn, " "); got != BrowsersTask+" ^build" {
		t.Errorf("a project without build does not depend on it: %q", got)
	}
	core := res.Projects["core"].Replace[TaskName]
	if got := strings.Join(core.Command, " "); got != "npm run e2e --silent" {
		t.Errorf("a library suite runs its own e2e script: %q", got)
	}
	browsers := res.Projects["core"].Replace[BrowsersTask]
	if got := strings.Join(browsers.Command, " "); got != "npx --no-install playwright install chromium" {
		t.Errorf("browsers %q", got)
	}
	if browsers.Cache {
		t.Error("installing browsers changes the machine, not files smith can restore: never cached")
	}
}

func TestContributeOnLinux(t *testing.T) {
	ws := t.TempDir()
	writeProject(t, ws, "packages/studio", `{"dependencies": {"playwright": "^1.63.0"}, "smith": {"playwright": {"display": true, "browsers": ["chromium", "firefox"]}}}`)
	projects := []protocol.Project{{Name: "studio", Root: "packages/studio", Tasks: map[string]protocol.Task{TaskName: e2eScript}}}

	onPlatform(t, "linux", true)
	res := Contribute(ws, protocol.ContributeParams{Projects: projects})
	if got := strings.Join(res.Projects["studio"].Replace[TaskName].Command, " "); got != "xvfb-run -a npm run e2e --silent" {
		t.Errorf("a display suite runs under xvfb on Linux: %q", got)
	}
	if got := strings.Join(res.Projects["studio"].Replace[BrowsersTask].Command, " "); got != "npx --no-install playwright install --with-deps chromium firefox" {
		t.Errorf("CI on Linux installs the system libraries too: %q", got)
	}

	onPlatform(t, "linux", false)
	res = Contribute(ws, protocol.ContributeParams{Projects: projects})
	if got := strings.Join(res.Projects["studio"].Replace[BrowsersTask].Command, " "); got != "npx --no-install playwright install chromium firefox" {
		t.Errorf("a developer's machine is not asked for root: %q", got)
	}

	onPlatform(t, "windows", true)
	res = Contribute(ws, protocol.ContributeParams{Projects: projects})
	if got := strings.Join(res.Projects["studio"].Replace[TaskName].Command, " "); got != "npm run e2e --silent" {
		t.Errorf("Windows has a display: %q", got)
	}
}

func TestForcedFlavour(t *testing.T) {
	ws := t.TempDir()
	writeProject(t, ws, "apps/desktop", `{"devDependencies": {"@playwright/test": "^1.63.0"}, "smith": {"playwright": {"flavour": "library"}}}`)
	if got := Detect(ws, "apps/desktop", true).Flavour; got != Library {
		t.Errorf("forced flavour %q", got)
	}
	if got := Detect(ws, "apps/desktop", false).Flavour; got != None {
		t.Errorf("a library suite needs its e2e script, forced or not: %q", got)
	}
}

func TestInCIKnowsAzurePipelines(t *testing.T) {
	t.Setenv("CI", "")
	t.Setenv("TF_BUILD", "")
	if inCI() {
		t.Fatal("a developer's machine")
	}
	t.Setenv("TF_BUILD", "True")
	if !inCI() {
		t.Error("Azure Pipelines sets TF_BUILD, not CI")
	}
	t.Setenv("TF_BUILD", "")
	t.Setenv("CI", "true")
	if !inCI() {
		t.Error("GitHub Actions sets CI")
	}
}
