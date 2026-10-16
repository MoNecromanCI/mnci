package dependencies

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/libs/plugin-python/venv"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestInstallAndUninstall(t *testing.T) {
	ws := t.TempDir()
	_ = os.WriteFile(filepath.Join(ws, venv.RequirementsFile), []byte("ruff\n"), 0o644)
	_ = os.MkdirAll(filepath.Join(ws, ".venv"), 0o755)
	_ = os.MkdirAll(filepath.Join(ws, "apps", "api"), 0o755)
	file := filepath.Join(ws, "apps", "api", "pyproject.toml")
	_ = os.WriteFile(file, []byte(pyproject), 0o644)
	var ran []string
	real := venv.Runner
	defer func() { venv.Runner = real }()
	venv.Runner = func(_ string, argv []string) error {
		ran = append(ran, strings.Join(argv[1:], " "))
		return nil
	}
	if _, err := os.Stat(venv.Path(ws)); err != nil {
		_ = os.MkdirAll(filepath.Dir(venv.Path(ws)), 0o755)
		_ = os.WriteFile(venv.Path(ws), nil, 0o755)
	}
	api := protocol.Project{Name: "api", Root: "apps/api"}
	if _, err := Install(ws, protocol.InstallParams{Project: &api, Packages: []string{"requests>=2.31"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(ws, protocol.InstallParams{Project: &api, Packages: []string{"pytest-cov"}, Dev: true}); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(file)
	if !strings.Contains(string(raw), `dependencies = ["requests>=2.31"]`) || !strings.Contains(string(raw), "[project.optional-dependencies]\ndev = [\"pytest-cov\"]") {
		t.Fatalf("pyproject:\n%s", raw)
	}
	joined := strings.Join(ran, "\n")
	if !strings.Contains(joined, "pip install --quiet --disable-pip-version-check -e ./apps/api") || !strings.Contains(joined, "--disable-pip-version-check pytest-cov") {
		t.Fatalf("ran:\n%s", joined)
	}
	res, err := Uninstall(ws, protocol.UninstallParams{Project: api, Packages: []string{"requests", "pytest_cov"}})
	if err != nil || len(res.Notes) != 1 {
		t.Fatal(err, res)
	}
	raw, _ = os.ReadFile(file)
	if strings.Contains(string(raw), "requests") || strings.Contains(string(raw), "pytest-cov") {
		t.Fatalf("pyproject:\n%s", raw)
	}
	if _, err := Uninstall(ws, protocol.UninstallParams{Project: api, Packages: []string{"flask"}}); err == nil || !strings.Contains(err.Error(), "does not require flask") {
		t.Fatalf("err %v", err)
	}
}
