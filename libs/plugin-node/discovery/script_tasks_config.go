package discovery

import (
	"strings"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// workspaceInputs are root files every Node task depends on.
var workspaceInputs = []string{
	"**/*",
	"{workspaceRoot}/package-lock.json",
	"{workspaceRoot}/tsconfig.base.json",
	"{workspaceRoot}/eslint.config.mjs",
}

// scriptTasks describes how each well-known script runs. Scripts not listed
// still become tasks, uncached and without dependencies.
var scriptTasks = map[string]protocol.Task{
	"build":     {Outputs: []string{"dist"}, DependsOn: []string{"^build"}, Cache: true, Description: "Compile to dist/"},
	"typecheck": {DependsOn: []string{"^build"}, Cache: true, Description: "Type-check without emitting"},
	"test":      {DependsOn: []string{"^build"}, Cache: true, Description: "Run unit tests"},
	"lint":      {DependsOn: []string{"^build"}, Cache: true, Description: "Lint and check formatting"},
	"start":     {DependsOn: []string{"build"}, Description: "Run the app"},
	"dev":       {Description: "Run the app in watch mode"},
}

// TaskFor builds the task that runs one package.json script through npm, so
// node_modules/.bin is on PATH exactly as when a developer runs it. A
// dependency on another script of the same project (start on build) is kept
// only when the project has that script: plain JavaScript has no build.
func TaskFor(script string, scripts map[string]string) protocol.Task {
	t, known := scriptTasks[script]
	var deps []string
	for _, d := range t.DependsOn {
		if _, has := scripts[d]; has || strings.HasPrefix(d, "^") {
			deps = append(deps, d)
		}
	}
	t.DependsOn = deps
	t.Command = []string{"npm", "run", script, "--silent"}
	if known && t.Cache {
		t.Inputs = workspaceInputs
	}
	if !known {
		t.Description = "npm script " + script
	}
	return t
}
