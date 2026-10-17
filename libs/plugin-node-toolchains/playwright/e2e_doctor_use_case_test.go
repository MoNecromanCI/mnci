package playwright

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func finding(t *testing.T, res protocol.DoctorResult, check string) protocol.Finding {
	t.Helper()
	for _, f := range res.Findings {
		if f.Check == check {
			return f
		}
	}
	t.Fatalf("no %q finding in %+v", check, res.Findings)
	return protocol.Finding{}
}

func TestDoctor(t *testing.T) {
	onPlatform(t, "linux", false)
	realLook := lookPath
	t.Cleanup(func() { lookPath = realLook })
	lookPath = func(string) (string, error) { return "", errors.New("not found") }

	ws := t.TempDir()
	if got := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{{Name: "x", Root: "x"}}}); len(got.Findings) != 0 {
		t.Fatalf("no suite, nothing to check: %+v", got.Findings)
	}

	writeProject(t, ws, "packages/studio", `{"dependencies": {"playwright": "^1.63.0"}, "smith": {"playwright": {"display": true}}}`)
	projects := []protocol.Project{{Name: "studio", Root: "packages/studio", Tasks: map[string]protocol.Task{TaskName: e2eScript}}}
	res := Doctor(ws, protocol.DoctorParams{Projects: projects})
	if f := finding(t, res, "Playwright is installed"); f.OK {
		t.Error("no node_modules: not installed")
	}
	if f := finding(t, res, "a virtual display for e2e suites that open windows"); f.OK || f.Remedy == "" {
		t.Errorf("display %+v", f)
	}
	if f := finding(t, res, "a pipeline runs the e2e tasks"); f.OK || !f.Warning {
		t.Errorf("pipeline %+v", f)
	}

	for _, f := range []string{"node_modules/playwright-core/package.json", ".github/workflows/e2e.yml"} {
		path := filepath.Join(ws, filepath.FromSlash(f))
		_ = os.MkdirAll(filepath.Dir(path), 0o755)
		_ = os.WriteFile(path, []byte("{}"), 0o644)
	}
	lookPath = func(string) (string, error) { return "/usr/bin/xvfb-run", nil }
	for _, f := range Doctor(ws, protocol.DoctorParams{Projects: projects}).Findings {
		if !f.OK {
			t.Errorf("all in place, yet %+v", f)
		}
	}
}

func TestDoctorFindsPlaywrightInTheSuitesOwnNodeModules(t *testing.T) {
	onPlatform(t, "darwin", false)
	ws := t.TempDir()
	writeProject(t, ws, "apps/web", `{"devDependencies": {"@playwright/test": "^1.63.0"}}`)
	local := filepath.Join(ws, "apps", "web", "node_modules", "playwright-core")
	_ = os.MkdirAll(local, 0o755)
	_ = os.WriteFile(filepath.Join(local, "package.json"), []byte("{}"), 0o644)
	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{{Name: "web", Root: "apps/web"}}})
	if f := finding(t, res, "Playwright is installed"); !f.OK {
		t.Error("a version npm could not hoist is installed beside the suite")
	}
}
