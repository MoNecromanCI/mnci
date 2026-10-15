package discovery

import (
	"github.com/MoNecromanCI/mnci/libs/plugin-python/venv"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

var inputs = []string{"**/*", "{workspaceRoot}/requirements-dev.txt", "{workspaceRoot}/ruff.toml"}

// Tasks returns the tasks of a Python project. Every project is installed in
// the venv in editable mode, so none needs another project built first. A
// project that configures mypy ([tool.mypy]) is also type-checked.
func Tasks(kind protocol.ProjectKind, importName string, mypy bool) map[string]protocol.Task {
	py := venv.Interpreter()
	tasks := map[string]protocol.Task{
		"lint": {Command: []string{py, "-m", "ruff", "check", "."}, Inputs: inputs, Cache: true, Description: "ruff check"},
		"test": {Command: []string{py, "-m", "pytest", "-q", "-p", "no:cacheprovider"}, Inputs: inputs, Cache: true, Description: "pytest"},
	}
	if mypy {
		tasks["typecheck"] = protocol.Task{Command: []string{py, "-m", "mypy", "."}, Inputs: inputs, Cache: true, Description: "mypy"}
	}
	switch kind {
	case protocol.ProjectPackage:
		tasks["build"] = protocol.Task{Command: []string{py, "-m", "build", "--outdir", "dist", "."}, Inputs: inputs, Outputs: []string{"dist"}, Cache: true, Description: "Build the wheel and sdist into dist/"}
	case protocol.ProjectApp:
		tasks["start"] = protocol.Task{Command: []string{py, "-m", importName}, Description: "Run the app"}
	}
	return tasks
}
