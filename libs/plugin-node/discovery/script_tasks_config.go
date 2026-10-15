package discovery

import "github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"

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
// node_modules/.bin is on PATH exactly as when a developer runs it.
func TaskFor(script string) protocol.Task {
	t, known := scriptTasks[script]
	t.Command = []string{"npm", "run", script, "--silent"}
	if known && t.Cache {
		t.Inputs = workspaceInputs
	}
	if !known {
		t.Description = "npm script " + script
	}
	return t
}
