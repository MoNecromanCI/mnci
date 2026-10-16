package discovery

import "github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"

// inputs are what a gem's results depend on: its sources and the
// workspace's resolution and RuboCop rules.
var inputs = []string{"**/*", "{workspaceRoot}/Gemfile", "{workspaceRoot}/Gemfile.lock", "{workspaceRoot}/.rubocop.yml"}

// Tasks are Bundler's for one gem, run in its directory: Bundler finds the
// workspace's Gemfile above it, so every gem shares one resolution. A gem
// does not build; publishing packs it into dist/.
func Tasks(kind protocol.ProjectKind, executable string, specs bool) map[string]protocol.Task {
	tasks := map[string]protocol.Task{
		"lint": {Command: []string{"bundle", "exec", "rubocop"}, Inputs: inputs, Cache: true, Description: "RuboCop (the workspace's .rubocop.yml)"},
	}
	if specs {
		tasks["test"] = protocol.Task{Command: []string{"bundle", "exec", "rspec"}, Inputs: inputs, Cache: true, Description: "RSpec"}
	}
	if kind == protocol.ProjectApp && executable != "" {
		tasks["start"] = protocol.Task{Command: []string{"bundle", "exec", "ruby", "exe/" + executable}, Description: "Run the app"}
	}
	return tasks
}
