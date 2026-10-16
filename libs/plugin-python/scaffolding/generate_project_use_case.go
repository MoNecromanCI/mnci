package scaffolding

import (
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-python/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

type templateData struct {
	Dir, Name, Distribution, Import, Kind, Stack string
}

// Generate scaffolds one Python project, creating the workspace venv and its
// tooling the first time, and installs the project in editable mode.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	distribution := distributionName(params.Workspace.Name, params.Name)
	data := templateData{Dir: params.Directory, Name: params.Name, Distribution: distribution, Import: discovery.ImportName(distribution), Kind: string(params.Kind), Stack: params.Stack}

	templates := maps.Clone(projectTemplates)
	switch params.Kind {
	case protocol.ProjectApp:
		maps.Copy(templates, appTemplates)
	case protocol.ProjectPackage:
		maps.Copy(templates, packageTemplates)
	}
	files, err := templating.Write(workspaceRoot, templates, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	if _, err := os.Stat(filepath.Join(workspaceRoot, "ruff.toml")); errors.Is(err, fs.ErrNotExist) {
		if err := os.WriteFile(filepath.Join(workspaceRoot, "ruff.toml"), []byte(ruffConfig), 0o644); err != nil {
			return protocol.GenerateResult{}, err
		}
		files = append(files, "ruff.toml")
	}
	created, err := venv.Ensure(workspaceRoot)
	if err != nil {
		return protocol.GenerateResult{Files: files}, err
	}
	files = append(files, created...)
	if err := venv.InstallEditable(workspaceRoot, params.Directory); err != nil {
		return protocol.GenerateResult{Files: files}, err
	}
	return protocol.GenerateResult{Files: files, Notes: []string{"Import it from another Python project as `import " + data.Import + "`."}}, nil
}

// distributionName prefixes the workspace name so packages are unique on a
// shared index: workspace "acme" and project "sdk" give "acme-sdk".
func distributionName(workspace, project string) string {
	ws := strings.ToLower(strings.TrimSpace(workspace))
	if ws == "" || strings.HasPrefix(project, ws+"-") {
		return project
	}
	return ws + "-" + project
}
