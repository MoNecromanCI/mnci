package scaffolding

import (
	"os"

	"github.com/MoNecromanCI/mnci/libs/plugin-python/venv"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Remove uninstalls a removed project's editable install from the venv, so
// nothing imports a package whose source is gone.
func Remove(workspaceRoot string, params protocol.RemoveParams) (protocol.RemoveResult, error) {
	name := params.Project.Name
	if len(params.Project.Exports) > 0 {
		name = params.Project.Exports[0]
	}
	if _, err := os.Stat(venv.Path(workspaceRoot)); err != nil {
		return protocol.RemoveResult{}, nil
	}
	note := "uninstalled " + name + " from .venv"
	if params.DryRun {
		note = "would uninstall " + name + " from .venv"
	} else if err := venv.Uninstall(workspaceRoot, name); err != nil {
		return protocol.RemoveResult{}, err
	}
	return protocol.RemoveResult{Notes: []string{note}}, nil
}
