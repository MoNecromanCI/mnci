package discovery

import (
	"path"
	"regexp"
	"strings"
)

// Manifest is the file that makes a directory a Bicep project.
const Manifest = "main.bicep"

// BicepFile is what smith reads from a main.bicep.
type BicepFile struct {
	// Version, Kind and Stack are the metadata version, smithKind and
	// smithStack statements.
	Version, Kind, Stack string
	// Modules are the local files its modules reference, relative to it.
	Modules []string
}

var (
	metadata = regexp.MustCompile(`(?m)^metadata\s+(version|smithKind|smithStack)\s*=\s*'([^']*)'`)
	module   = regexp.MustCompile(`(?m)^\s*module\s+\w+\s+'([^']+)'`)
)

// ParseBicep reads a main.bicep. A module from a registry (br:, ts:) is not
// a workspace file.
func ParseBicep(src string) BicepFile {
	var f BicepFile
	for _, m := range metadata.FindAllStringSubmatch(src, -1) {
		switch m[1] {
		case "version":
			f.Version = m[2]
		case "smithKind":
			f.Kind = m[2]
		case "smithStack":
			f.Stack = m[2]
		}
	}
	for _, m := range module.FindAllStringSubmatch(src, -1) {
		if !strings.Contains(m[1], ":") {
			f.Modules = append(f.Modules, m[1])
		}
	}
	return f
}

// ModuleRoot is the workspace directory of a module file referenced from
// the project in root, or empty when it leaves the workspace.
func ModuleRoot(root, file string) string {
	dir := path.Dir(path.Join(root, file))
	if dir == ".." || strings.HasPrefix(dir, "../") {
		return ""
	}
	return dir
}
