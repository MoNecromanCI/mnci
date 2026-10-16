package discovery

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// inputs are what a package's results depend on: its sources and the
// workspace's resolution.
var inputs = []string{"**/*", "!vendor/**", "{workspaceRoot}/composer.json", "{workspaceRoot}/composer.lock"}

// Shape is what a package's tasks depend on.
type Shape struct {
	Kind protocol.ProjectKind
	// Tests is true when the package has a tests/ directory.
	Tests bool
	// Bin is the app's executable, project-relative.
	Bin string
}

// Tasks are PHPUnit's and PHPStan's for one package, run in its directory
// with the workspace's vendor/: one resolution, one autoloader, for every
// package. PHPStan runs at level 8 over the sources and the tests.
func Tasks(s Shape) map[string]protocol.Task {
	vendor := "{workspaceRoot}/vendor/"
	paths := []string{"src"}
	if s.Tests {
		paths = append(paths, "tests")
	}
	lint := append([]string{"php", vendor + "bin/phpstan", "analyse", "--no-progress", "--memory-limit=1G", "--level", "8", "--autoload-file", vendor + "autoload.php"}, paths...)
	tasks := map[string]protocol.Task{
		"lint": {Command: lint, Inputs: inputs, Cache: true, Description: "PHPStan, level 8"},
	}
	if s.Tests {
		tasks["test"] = protocol.Task{Command: []string{"php", vendor + "bin/phpunit", "--bootstrap", vendor + "autoload.php", "tests"}, Inputs: inputs, Cache: true, Description: "PHPUnit"}
	}
	if s.Kind == protocol.ProjectApp && s.Bin != "" {
		tasks["start"] = protocol.Task{Command: []string{"php", s.Bin}, Description: "Run the app"}
	}
	return tasks
}
