package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-vertical-slices/languages"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TaskName is the task the convention adds to each project.
const TaskName = "slice-check"

// sliceCheck runs smith's own checker on the project root. Only file and
// folder names matter, but a new or renamed file changes the inputs, so the
// cached result is reused exactly when the layout did not change.
func sliceCheck(rules languages.Rules) protocol.Task {
	return protocol.Task{
		Command:     []string{"{smith}", "slices", "check", "--lang", rules.Name, "."},
		Inputs:      rules.Inputs,
		Cache:       true,
		Description: "checks the vertical-slice layout: role-named files, no junk-drawer folders",
	}
}
