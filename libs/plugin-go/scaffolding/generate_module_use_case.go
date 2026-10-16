package scaffolding

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/templating"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/toolrun"
)

// GoRunner runs the go command; tests replace it.
var GoRunner = func(dir string, argv []string) error { return toolrun.Run(dir, argv) }

type templateData struct {
	Dir, Name, Module, Package string
}

// Generate scaffolds one Go module and adds it to go.work (creating go.work the
// first time).
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))
	if _, err := os.Stat(dir); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	prefix := params.Workspace.ModulePrefix
	if prefix == "" {
		prefix = "example.com/" + strings.ToLower(params.Workspace.Name)
	}
	data := templateData{Dir: params.Directory, Name: params.Name, Module: prefix + "/" + params.Directory, Package: PackageName(params.Name)}

	var files []string
	if _, err := os.Stat(filepath.Join(workspaceRoot, "go.work")); errors.Is(err, fs.ErrNotExist) {
		if err := GoRunner(workspaceRoot, []string{"go", "work", "init"}); err != nil {
			return protocol.GenerateResult{}, err
		}
		files = append(files, "go.work")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return protocol.GenerateResult{}, err
	}
	if err := GoRunner(dir, []string{"go", "mod", "init", data.Module}); err != nil {
		return protocol.GenerateResult{}, err
	}
	if err := declareKind(filepath.Join(dir, "go.mod"), params.Kind, params.Stack); err != nil {
		return protocol.GenerateResult{}, err
	}
	files = append(files, params.Directory+"/go.mod")

	templates := maps.Clone(sliceTemplates)
	if params.Kind == protocol.ProjectApp {
		maps.Copy(templates, appTemplates)
	} else {
		maps.Copy(templates, libTemplates)
	}
	written, err := templating.Write(workspaceRoot, templates, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	files = append(files, written...)
	if err := GoRunner(workspaceRoot, []string{"go", "work", "use", "./" + params.Directory}); err != nil {
		return protocol.GenerateResult{}, err
	}
	return protocol.GenerateResult{Files: files, Notes: []string{"Import it from another Go project as " + data.Module + "/<slice>."}}, nil
}

// declareKind records the kind, and the stack when there is one, as the
// first lines of go.mod: comments the Go toolchain keeps.
func declareKind(goMod string, kind protocol.ProjectKind, stack string) error {
	src, err := os.ReadFile(goMod)
	if err != nil {
		return err
	}
	head := "// mnci:kind " + string(kind) + "\n"
	if stack != "" {
		head += "// mnci:stack " + stack + "\n"
	}
	return os.WriteFile(goMod, append([]byte(head), src...), 0o644)
}

// PackageName turns a project name into a valid Go package name.
func PackageName(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" || unicode.IsDigit(rune(out[0])) {
		out = "pkg" + out
	}
	return out
}
