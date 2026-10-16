package discovery

import (
	"runtime"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

var inputs = []string{"**/*", "{workspaceRoot}/go.work", "{workspaceRoot}/go.work.sum"}

// NativeTag marks a module built on each OS (cgo): mnci ci verify leaves it
// to mnci ci native, which runs on every runner.
const NativeTag = "native"

// Tasks returns the tasks of a Go module. Go builds its dependencies itself,
// so no task depends on another project's build; the cache key still follows
// every dependency's sources. A native module builds with cgo.
func Tasks(name string, kind protocol.ProjectKind, native bool) map[string]protocol.Task {
	tasks := map[string]protocol.Task{
		"test": {Command: []string{"go", "test", "./..."}, Inputs: inputs, Cache: true, Description: "go test ./..."},
		"lint": {Command: []string{"go", "vet", "./..."}, Inputs: inputs, Cache: true, Description: "go vet ./..."},
	}
	if kind == protocol.ProjectApp {
		binary := name
		if runtime.GOOS == "windows" {
			binary += ".exe"
		}
		tasks["build"] = protocol.Task{Command: []string{"go", "build", "-trimpath", "-o", "dist/" + binary, "."}, Inputs: inputs, Outputs: []string{"dist"}, Cache: true, Description: "Build the binary into dist/"}
		tasks["start"] = protocol.Task{Command: []string{"go", "run", "."}, Description: "Run the app"}
		tasks["package"] = protocol.Task{
			Command:     []string{"go", "build", "-trimpath", "-o", "{workspaceRoot}/dist/drop/" + name + "/" + binary, "."},
			Description: "Build the binary into dist/drop/" + name + "/ (mnci ci pack)",
		}
	} else {
		tasks["build"] = protocol.Task{Command: []string{"go", "build", "./..."}, Inputs: inputs, Cache: true, Description: "Compile every package"}
	}
	if native {
		for n, t := range tasks {
			t.Env = map[string]string{"CGO_ENABLED": "1"}
			tasks[n] = t
		}
	}
	return tasks
}
