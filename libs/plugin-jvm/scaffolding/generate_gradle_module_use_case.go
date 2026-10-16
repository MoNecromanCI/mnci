package scaffolding

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/MoNecromanCI/mnci/libs/plugin-jvm/discovery"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/templating"
)

// gradleState is what Gradle and the Kotlin compiler keep in the workspace
// root: caches, never sources.
var gradleState = []string{".gradle", ".kotlin"}

// GenerateGradle writes a Kotlin module, includes it in the settings, and
// writes the settings and root script the first time.
func GenerateGradle(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))
	if _, err := os.Stat(dir); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	settingsFile := filepath.Join(workspaceRoot, settings)
	workspace := params.Workspace.Name
	templates := maps.Clone(kotlinSliceTemplates)
	maps.Copy(templates, gradleModuleTemplates)
	if params.Kind == protocol.ProjectApp {
		maps.Copy(templates, kotlinAppTemplates)
	}
	src, err := os.ReadFile(settingsFile)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		maps.Copy(templates, gradleRootTemplates)
	case err != nil:
		return protocol.GenerateResult{}, err
	default:
		if name := RootProjectName(string(src)); name != "" {
			workspace = name
		}
	}
	data := templateData{
		Dir: params.Directory, Name: params.Name, Workspace: workspace, Group: GroupID(workspace),
		ArtifactID: workspace + "-" + params.Name, Kind: string(params.Kind), Stack: params.Stack,
	}
	data.Package = data.Group + "." + identifier(params.Name)
	data.PackagePath = strings.ReplaceAll(data.Package, ".", "/")
	files, err := templating.Write(workspaceRoot, templates, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	if src, err = os.ReadFile(settingsFile); err != nil {
		return protocol.GenerateResult{}, err
	}
	path := discovery.ProjectPath(params.Directory)
	if out, changed := AddInclude(string(src), path); changed {
		if err := os.WriteFile(settingsFile, []byte(out), 0o644); err != nil {
			return protocol.GenerateResult{}, err
		}
		if !slices.Contains(files, settings) {
			files = append(files, settings)
		}
	}
	if added, err := ignoreGradleState(workspaceRoot); err != nil {
		return protocol.GenerateResult{}, err
	} else if added {
		files = append(files, ".gitignore")
	}
	return protocol.GenerateResult{Files: files, Notes: []string{"Depend on it from another module with implementation(project(\"" + path + "\"))."}}, nil
}

// ignoreGradleState adds Gradle's and Kotlin's workspace caches to the
// .gitignore, so they never count as sources.
func ignoreGradleState(workspaceRoot string) (bool, error) {
	file := filepath.Join(workspaceRoot, ".gitignore")
	raw, err := os.ReadFile(file)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	ignored := map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		ignored[strings.Trim(strings.TrimSpace(line), "/")] = true
	}
	var missing []string
	for _, dir := range gradleState {
		if !ignored[dir] {
			missing = append(missing, dir+"/")
		}
	}
	if len(missing) == 0 {
		return false, nil
	}
	text := string(raw)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return true, os.WriteFile(file, []byte(text+"\n# Gradle\n"+strings.Join(missing, "\n")+"\n"), 0o644)
}
