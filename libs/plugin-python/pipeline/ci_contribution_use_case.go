package pipeline

import (
	"os"
	"path/filepath"

	"github.com/MoNecromanCI/mnci/libs/plugin-python/venv"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// CI returns the Python contribution. Setup steps are computed on the machine
// that runs them, so the interpreter path matches that OS.
func CI(params protocol.CIParams) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	py := venv.Interpreter()
	launcher := venv.Launchers()[0]
	setup := []protocol.Step{
		{Name: "Create the Python venv", Command: append(append([]string{}, launcher...), "-m", "venv", "--upgrade-deps", ".venv"), Blocking: true},
		{Name: "Install the Python toolchain", Command: []string{py, "-m", "pip", "install", "--quiet", "--disable-pip-version-check", "-r", venv.RequirementsFile}, Blocking: true},
	}
	for _, p := range params.Projects {
		setup = append(setup, protocol.Step{Name: "Install " + p.Name, Command: []string{py, "-m", "pip", "install", "--quiet", "--disable-pip-version-check", "-e", "./" + p.Root}, Blocking: true})
	}
	result := protocol.CIResult{
		Toolchains: []protocol.Toolchain{runtime(venv.PythonVersion)},
		Setup:      setup,
		Audit:      []protocol.Step{{Name: "pip-audit", Command: []string{py, "-m", "pip_audit", "--skip-editable"}, Blocking: false}},
	}
	for _, p := range params.Projects {
		if p.Kind == protocol.ProjectPackage {
			result.Credentials = []protocol.Credential{{Env: "PYPI_TOKEN", Description: "PyPI API token (or set TWINE_USERNAME/TWINE_PASSWORD and TWINE_REPOSITORY_URL)", Required: true}}
			break
		}
	}
	return result
}

// Doctor checks the venv and the requirements file exist.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	if len(params.Projects) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	_, venvErr := os.Stat(venv.Path(workspaceRoot))
	_, reqErr := os.Stat(filepath.Join(workspaceRoot, venv.RequirementsFile))
	return protocol.DoctorResult{Findings: []protocol.Finding{
		{Check: "Python venv exists", OK: venvErr == nil, Detail: ".venv is missing at the workspace root", Remedy: "run `mnci ci setup` (or `python -m venv .venv` and install requirements-dev.txt)"},
		{Check: "requirements-dev.txt exists", OK: reqErr == nil, Detail: "the Python toolchain file is missing", Remedy: "run `mnci add` for a Python project, or restore requirements-dev.txt"},
	}}
}
