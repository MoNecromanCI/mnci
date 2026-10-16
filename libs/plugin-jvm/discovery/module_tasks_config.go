package discovery

import (
	"strings"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// inputs are what a module's results depend on: its sources and pom, and the
// workspace's root pom (plugin versions, Checkstyle rules).
var inputs = []string{"src/**", "pom.xml", "{workspaceRoot}/pom.xml", "{workspaceRoot}/.mvn/**"}

// BuildDir is where a module builds (the root pom's <directory>): the
// workspace's dist/maven/<artifactId>.
func BuildDir(artifactID string) string { return protocol.BuildDir + "/maven/" + artifactID }

// Tasks are Maven's for one module, run from the workspace root (-pl). Each
// touches only its own module: build installs it into the local repository
// after the modules it depends on (^build), so the modules that use it
// resolve it there; test and start wait for the module's build; lint is
// Checkstyle with the root pom's rules.
func Tasks(root, artifactID string, kind protocol.ProjectKind) map[string]protocol.Task {
	mvn := func(args ...string) []string { return append([]string{"mvn", "-B", "-q", "-pl", root}, args...) }
	up := strings.Repeat("../", strings.Count(root, "/")+1)
	tasks := map[string]protocol.Task{
		"build": {Command: mvn("install", "-DskipTests"), Cwd: ".", Inputs: inputs, Outputs: []string{up + BuildDir(artifactID)}, DependsOn: []string{"^build"}, Description: "mvn install (into the local repository, for the modules that use it)"},
		"test":  {Command: mvn("test"), Cwd: ".", Inputs: inputs, DependsOn: []string{"build"}, Cache: true, Description: "mvn test"},
		"lint":  {Command: mvn("checkstyle:check"), Cwd: ".", Inputs: inputs, Cache: true, Description: "Checkstyle (the root pom's rules)"},
	}
	if kind == protocol.ProjectApp {
		tasks["start"] = protocol.Task{Command: mvn("exec:java"), Cwd: ".", DependsOn: []string{"build"}, Description: "Run the app (exec:java)"}
	}
	return tasks
}
