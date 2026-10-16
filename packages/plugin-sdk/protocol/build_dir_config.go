package protocol

import "strings"

// BuildDir is the directory at the workspace root every build writes into:
// smith's opinion, so all output is in one place, outside every source tree,
// and cleaning is deleting one folder. A plugin builds a project into
// ProjectBuildDir; smith doctor fails a task whose outputs fall outside it.
const BuildDir = "dist"

// ProjectBuildDir is where a project's build goes, workspace-relative:
// dist/<project dir>, mirroring the workspace (dist/apps/web).
func ProjectBuildDir(root string) string { return BuildDir + "/" + strings.Trim(root, "/") }

// BuildDirFromProject is the same directory relative to the project
// (../../dist/apps/web): what a tool configured from inside the project, or a
// task's outputs, name.
func BuildDirFromProject(root string) string {
	root = strings.Trim(root, "/")
	return strings.Repeat("../", strings.Count(root, "/")+1) + ProjectBuildDir(root)
}

// InBuildDir reports whether a workspace-relative path is inside BuildDir.
func InBuildDir(path string) bool {
	return path == BuildDir || strings.HasPrefix(path, BuildDir+"/")
}
