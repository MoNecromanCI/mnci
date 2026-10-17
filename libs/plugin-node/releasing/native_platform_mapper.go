package releasing

import "fmt"

// vscodeTargets are the Marketplace's platform names for Go's os/arch.
var vscodeTargets = map[string]string{
	"windows/amd64": "win32-x64",
	"windows/arm64": "win32-arm64",
	"linux/amd64":   "linux-x64",
	"linux/arm64":   "linux-arm64",
	"darwin/amd64":  "darwin-x64",
	"darwin/arm64":  "darwin-arm64",
}

// VSCodeTarget is the `vsce --target` of a runner's platform ("os/arch" in
// Go's spelling), for an extension that ships a build per platform.
func VSCodeTarget(platform string) (string, error) {
	if t, ok := vscodeTargets[platform]; ok {
		return t, nil
	}
	return "", fmt.Errorf("no VS Code target for %s: the Marketplace takes win32, linux and darwin on x64 or arm64", platform)
}
