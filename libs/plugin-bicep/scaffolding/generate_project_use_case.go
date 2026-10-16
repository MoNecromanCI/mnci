package scaffolding

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

type templateData struct {
	Dir, Name, Kind, Stack string
	// Storage is a storage account name suffix: lowercase letters and
	// digits only, at most 22 characters after "st".
	Storage string
}

// Generate writes a Bicep project. A deployment (app) gets parameters; a
// module others use (lib) does not. Bicep modules are not published to a
// registry yet, so there is no package.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if params.Kind == protocol.ProjectPackage {
		return protocol.GenerateResult{}, errors.New("a Bicep project is an app (a deployment) or a lib (a module): publishing to a Bicep registry is not supported yet")
	}
	if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	files := maps.Clone(templates)
	if params.Kind == protocol.ProjectApp {
		maps.Copy(files, appTemplates)
	}
	data := templateData{Dir: params.Directory, Name: params.Name, Kind: string(params.Kind), Stack: params.Stack, Storage: storageSuffix(params.Workspace.Name + params.Name)}
	written, err := templating.Write(workspaceRoot, files, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	out := protocol.ProjectBuildDir(params.Directory)
	note := "After smith build, deploy it with az deployment group create --template-file " + out + "/template/main.json --parameters " + out + "/parameters/main.parameters.json."
	if params.Kind == protocol.ProjectLib {
		note = "Use it from a deployment: module " + params.Name + " '<relative path>/" + params.Directory + "/main.bicep' = { ... }."
	}
	return protocol.GenerateResult{Files: written, Notes: []string{note}}, nil
}

func storageSuffix(name string) string {
	var b strings.Builder
	for _, c := range strings.ToLower(name) {
		if b.Len() < 22 && (unicode.IsDigit(c) || (c >= 'a' && c <= 'z')) {
			b.WriteRune(c)
		}
	}
	return b.String()
}
