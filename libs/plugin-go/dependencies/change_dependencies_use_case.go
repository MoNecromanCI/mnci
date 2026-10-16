package dependencies

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/toolrun"
)

// GoRunner runs the go command; tests replace it.
var GoRunner = func(dir string, argv []string) error { return toolrun.Run(dir, argv) }

func moduleFiles(root string) []string { return []string{root + "/go.mod", root + "/go.sum"} }

// Install requires modules in one project's go.mod; Dev records them as
// tools (go get -tool), Go's development dependencies. Without a project it
// downloads every module's requirements.
func Install(workspaceRoot string, params protocol.InstallParams) (protocol.DependencyResult, error) {
	if params.Project == nil {
		for _, p := range params.Projects {
			if err := GoRunner(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root)), []string{"go", "mod", "download"}); err != nil {
				return protocol.DependencyResult{}, err
			}
		}
		return protocol.DependencyResult{}, nil
	}
	if len(params.Packages) == 0 {
		return protocol.DependencyResult{}, GoRunner(dir(workspaceRoot, *params.Project), []string{"go", "mod", "download"})
	}
	argv := []string{"go", "get"}
	if params.Dev {
		argv = append(argv, "-tool")
	}
	if err := GoRunner(dir(workspaceRoot, *params.Project), append(argv, params.Packages...)); err != nil {
		return protocol.DependencyResult{}, err
	}
	res := protocol.DependencyResult{Files: moduleFiles(params.Project.Root)}
	if !params.Dev {
		res.Notes = []string{"go mod tidy drops a requirement nothing imports yet: import it before tidying"}
	}
	return res, nil
}

// Uninstall drops modules from one project's go.mod (go get <module>@none),
// as tools when go.mod lists them as tools.
func Uninstall(workspaceRoot string, params protocol.UninstallParams) (protocol.DependencyResult, error) {
	d := dir(workspaceRoot, params.Project)
	gomod, err := os.ReadFile(filepath.Join(d, "go.mod"))
	if err != nil {
		return protocol.DependencyResult{}, err
	}
	var tools, modules []string
	for _, pkg := range params.Packages {
		path := strings.SplitN(pkg, "@", 2)[0]
		if isTool(string(gomod), path) {
			tools = append(tools, path+"@none")
		} else {
			modules = append(modules, path+"@none")
		}
	}
	if len(tools) > 0 {
		if err := GoRunner(d, append([]string{"go", "get", "-tool"}, tools...)); err != nil {
			return protocol.DependencyResult{}, err
		}
	}
	if len(modules) > 0 {
		if err := GoRunner(d, append([]string{"go", "get"}, modules...)); err != nil {
			return protocol.DependencyResult{}, err
		}
	}
	return protocol.DependencyResult{Files: moduleFiles(params.Project.Root)}, nil
}

func dir(workspaceRoot string, p protocol.Project) string {
	return filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
}

// isTool reports whether go.mod has a tool directive for path, alone or in a
// tool ( ... ) block.
func isTool(gomod, path string) bool {
	q := regexp.QuoteMeta(path)
	return regexp.MustCompile(`(?m)^tool[ \t]+`+q+`[ \t]*(//.*)?\r?$`).MatchString(gomod) ||
		regexp.MustCompile(`(?ms)^tool[ \t]*\([^)]*^[ \t]+`+q+`[ \t]*(//[^\n]*)?\r?$`).MatchString(gomod)
}
