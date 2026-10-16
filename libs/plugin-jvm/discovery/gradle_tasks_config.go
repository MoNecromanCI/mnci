package discovery

import (
	"strings"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// gradleInputs are what a Gradle module's results depend on: its sources,
// script and properties, and the workspace's settings, root script,
// properties and version catalog.
var gradleInputs = []string{
	"src/**", "build.gradle.kts", "build.gradle", "gradle.properties",
	"{workspaceRoot}/settings.gradle.kts", "{workspaceRoot}/settings.gradle",
	"{workspaceRoot}/build.gradle.kts", "{workspaceRoot}/build.gradle",
	"{workspaceRoot}/gradle.properties", "{workspaceRoot}/gradle/**",
}

// GradleTasks are Gradle's for one module, run from the workspace root and
// selected by its project path. build assembles it after the modules it
// depends on (^build); test and start wait for it; lint is what the
// module's plugins add (ktlint for Kotlin, Checkstyle for Java).
func GradleTasks(root string, b GradleBuild, kind protocol.ProjectKind) map[string]protocol.Task {
	gradle := func(tasks ...string) []string {
		argv := []string{"gradle", "--console=plain", "-q"}
		for _, t := range tasks {
			argv = append(argv, b.Path+":"+t)
		}
		return argv
	}
	build := protocol.Task{Command: gradle("assemble"), Cwd: ".", Inputs: gradleInputs, DependsOn: []string{"^build"}, Description: "gradle assemble"}
	if b.BuildDir != "" {
		build.Outputs = []string{strings.Repeat("../", strings.Count(root, "/")+1) + b.BuildDir}
	}
	tasks := map[string]protocol.Task{
		"build": build,
		"test":  {Command: gradle("test"), Cwd: ".", Inputs: gradleInputs, DependsOn: []string{"build"}, Cache: true, Description: "gradle test"},
	}
	if len(b.Lint) > 0 {
		tasks["lint"] = protocol.Task{Command: gradle(b.Lint...), Cwd: ".", Inputs: gradleInputs, Cache: true, Description: "gradle " + strings.Join(b.Lint, " ")}
	}
	if kind == protocol.ProjectApp {
		tasks["start"] = protocol.Task{Command: gradle("run"), Cwd: ".", DependsOn: []string{"build"}, Description: "Run the app (gradle run)"}
	}
	return tasks
}
