package pipeline

import "github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"

// runtime installs Python in the generated pipelines.
func runtime(version string) protocol.Toolchain {
	return protocol.Toolchain{
		Name: "python", Version: version,
		GitHub: &protocol.ActionStep{Uses: "actions/setup-python@v5", Inputs: []protocol.Input{{Name: "python-version", Value: version}}},
		Azure:  &protocol.ActionStep{Uses: "UsePythonVersion@0", Inputs: []protocol.Input{{Name: "versionSpec", Value: version}}},
	}
}
