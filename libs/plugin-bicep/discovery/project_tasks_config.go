package discovery

import (
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// inputs are what a project's results depend on: its Bicep files, its
// parameters and its linter rules.
var inputs = []string{"**/*.bicep", "**/*.bicepparam", "bicepconfig.json"}

// Tasks are the Azure CLI's for one project, run in it. lint fails on the
// rules bicepconfig.json raises to errors (az bicep lint exits 0 on
// warnings); build compiles main.bicep into the workspace's
// dist/<project dir>/template, after params compiles main.bicepparam, when
// there is one, into dist/<project dir>/parameters. Each owns its own
// directory: mnci empties a task's outputs before it runs.
func Tasks(root string, params bool) map[string]protocol.Task {
	out := protocol.BuildDirFromProject(root)
	build := protocol.Task{Command: []string{"az", "bicep", "build", "--file", Manifest, "--outdir", out + "/template"}, Inputs: inputs, Outputs: []string{out + "/template"}, Cache: true, Description: "Compile main.bicep to an ARM template in dist/"}
	tasks := map[string]protocol.Task{
		"lint": {Command: []string{"az", "bicep", "lint", "--file", Manifest}, Inputs: inputs, Cache: true, Description: "az bicep lint (bicepconfig.json's rules)"},
	}
	if params {
		tasks["params"] = protocol.Task{Command: []string{"az", "bicep", "build-params", "--file", "main.bicepparam", "--outfile", out + "/parameters/main.parameters.json"}, Inputs: inputs, Outputs: []string{out + "/parameters"}, Cache: true, Description: "Compile main.bicepparam to a parameters file in dist/"}
		build.DependsOn = []string{"params"}
	}
	tasks["build"] = build
	return tasks
}
