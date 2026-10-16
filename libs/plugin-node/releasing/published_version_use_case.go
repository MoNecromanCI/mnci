package releasing

import (
	"strings"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/toolrun"
)

// capture runs npm and returns stdout and stderr; tests replace it.
var capture = func(dir string, argv []string) (string, string, error) {
	return toolrun.Capture(dir, argv)
}

// Published asks the registry .npmrc points at for the newest version of a
// package. A package the registry does not know (E404) is not published yet;
// any other failure means the registry could not be asked.
func Published(workspaceRoot string, params protocol.PublishedParams) (protocol.PublishedResult, error) {
	if len(params.Project.Exports) == 0 {
		return protocol.PublishedResult{}, nil
	}
	name := params.Project.Exports[0]
	registry, _, _ := capture(workspaceRoot, []string{"npm", "config", "get", "registry"})
	result := protocol.PublishedResult{Registry: strings.TrimSpace(registry)}
	// npm refuses view inside a workspace package, so ask at the root.
	out, errOut, err := capture(workspaceRoot, []string{"npm", "view", name, "version", "--fetch-retries=0", "--fetch-timeout=15000"})
	switch {
	case err == nil:
		result.Version = strings.TrimSpace(out)
	case strings.Contains(errOut, "E404") || strings.Contains(errOut, "404 Not Found"):
	default:
		result.Unreachable = true
		result.Detail = firstLine(errOut, err.Error())
	}
	return result, nil
}

func firstLine(text, fallback string) string {
	for _, line := range strings.Split(text, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return fallback
}
