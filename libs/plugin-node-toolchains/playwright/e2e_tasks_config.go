package playwright

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TaskName is the task the plugin owns.
const TaskName = "e2e"

// BrowsersTask installs the browsers a project's e2e task needs. It is a task
// of its own, so `smith run e2e` installs them wherever it runs, and nothing
// else (lint, test, build, smith ci setup) ever downloads a browser.
const BrowsersTask = "e2e-browsers"

// inputs are what an e2e run depends on besides the project's own files.
var inputs = []string{"**/*", "{workspaceRoot}/package-lock.json"}

// ReportsDir is where a project's e2e reports and traces go, workspace-relative:
// dist/e2e/<project dir>. Not inside dist/<project dir>, which is the build's
// output and is replaced whenever the build is restored from the cache.
func ReportsDir(root string) string {
	return protocol.BuildDir + "/" + TaskName + "/" + strings.Trim(root, "/")
}

// ReportsDirFromProject is the same directory relative to the project
// (../../dist/e2e/apps/web), as the task's outputs and a Playwright config name it.
func ReportsDirFromProject(root string) string {
	return strings.Repeat("../", strings.Count(strings.Trim(root, "/"), "/")+1) + ReportsDir(root)
}

// platform is the OS tasks are built for; tests replace it.
var platform = runtime.GOOS

// inCI reports whether smith runs in a CI job; tests replace it.
var inCI = func() bool { return os.Getenv("CI") != "" }

// Contribute gives every project with a Playwright suite its e2e task and
// the task that installs its browsers first.
func Contribute(workspaceRoot string, params protocol.ContributeParams) protocol.ContributeResult {
	res := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, p := range params.Projects {
		script, hasScript := p.Tasks[TaskName]
		suite := Detect(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root)), hasScript)
		if suite.Flavour == None {
			continue
		}
		res.Projects[p.Name] = protocol.Contribution{Replace: map[string]protocol.Task{
			TaskName:     e2eTask(p, suite, script, hasScript),
			BrowsersTask: browsersTask(suite),
		}}
	}
	return res
}

// e2eTask runs the project's e2e script when it has one (the library flavour
// always does), else the Playwright test runner. It runs after the browsers
// are installed and after the builds it tests, and is cached like test, its
// reports with it.
func e2eTask(p protocol.Project, suite Suite, script protocol.Task, hasScript bool) protocol.Task {
	command := []string{"npx", "--no-install", "playwright", "test"}
	if hasScript {
		command = script.Command
	}
	if suite.Options.Display && platform == "linux" {
		command = append([]string{"xvfb-run", "-a"}, command...)
	}
	deps := []string{BrowsersTask, "^build"}
	if _, ok := p.Tasks["build"]; ok {
		deps = append(deps, "build")
	}
	return protocol.Task{
		Command:     command,
		Inputs:      inputs,
		Outputs:     []string{ReportsDirFromProject(p.Root)},
		DependsOn:   deps,
		Cache:       true,
		Description: "Run the end-to-end tests (Playwright " + string(suite.Flavour) + ", " + strings.Join(suite.BrowserList(), ", ") + ")",
	}
}

// browsersTask installs the browsers at the build the project's Playwright
// version pins. In CI on Linux it also installs their system libraries,
// which needs root; a developer's machine is left to `smith doctor`.
func browsersTask(suite Suite) protocol.Task {
	command := []string{"npx", "--no-install", "playwright", "install"}
	if platform == "linux" && inCI() {
		command = append(command, "--with-deps")
	}
	return protocol.Task{
		Command:     append(command, suite.BrowserList()...),
		Description: "Install the browsers the e2e task needs (" + strings.Join(suite.BrowserList(), ", ") + ")",
	}
}
