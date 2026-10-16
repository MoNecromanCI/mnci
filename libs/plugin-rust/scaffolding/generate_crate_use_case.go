package scaffolding

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/MoNecromanCI/mnci/libs/plugin-rust/discovery"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/templating"
)

type templateData struct {
	Dir, Name, Crate, Kind, Stack string
}

// CrateName is a project's crate name: the workspace's name first, so a
// published crate does not clash on crates.io (acme-core).
func CrateName(workspace, name string) string {
	if workspace == "" || strings.HasPrefix(name, workspace+"-") {
		return name
	}
	return workspace + "-" + name
}

// Generate writes a crate, lists it in the root Cargo.toml's workspace
// members (creating the workspace the first time), and keeps target/ out of
// git.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))
	if _, err := os.Stat(dir); err == nil {
		return protocol.GenerateResult{}, fmt.Errorf("%s already exists", params.Directory)
	}
	rootManifest := filepath.Join(workspaceRoot, "Cargo.toml")
	var files []string
	src, err := os.ReadFile(rootManifest)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		src = []byte(NewWorkspace(params.Directory))
	case err != nil:
		return protocol.GenerateResult{}, err
	default:
		if crate, err := discovery.ParseCargo(src); err == nil && crate.Name != "" {
			return protocol.GenerateResult{}, fmt.Errorf("the root Cargo.toml is a crate (%s), not a workspace: make it a [workspace] first", crate.Name)
		}
		updated, err := AddMember(string(src), params.Directory)
		if err != nil {
			return protocol.GenerateResult{}, err
		}
		src = []byte(updated)
	}
	templates := maps.Clone(sliceTemplates)
	maps.Copy(templates, manifestTemplate)
	if params.Kind == protocol.ProjectApp {
		maps.Copy(templates, appTemplates)
	} else {
		maps.Copy(templates, libTemplates)
	}
	data := templateData{Dir: params.Directory, Name: params.Name, Crate: CrateName(params.Workspace.Name, params.Name), Kind: string(params.Kind), Stack: params.Stack}
	written, err := templating.Write(workspaceRoot, templates, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	files = append(files, written...)
	if err := os.WriteFile(rootManifest, src, 0o644); err != nil {
		return protocol.GenerateResult{}, err
	}
	files = append(files, "Cargo.toml")
	if added, err := ignoreTarget(workspaceRoot); err != nil {
		return protocol.GenerateResult{}, err
	} else if added {
		files = append(files, ".gitignore")
	}
	return protocol.GenerateResult{Files: files, Notes: []string{"Import it from another crate as " + data.Crate + " = { path = \"<relative path>\" }."}}, nil
}

// ignoreTarget adds Cargo's target/ to the workspace's .gitignore.
func ignoreTarget(workspaceRoot string) (bool, error) {
	path := filepath.Join(workspaceRoot, ".gitignore")
	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if l := strings.TrimSpace(line); l == "target/" || l == "/target" || l == "/target/" || l == "target" {
			return false, nil
		}
	}
	text := string(raw)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return true, os.WriteFile(path, []byte(text+"\n# Rust\ntarget/\n"), 0o644)
}

// Remove takes a removed crate out of the workspace's members.
func Remove(workspaceRoot string, params protocol.RemoveParams) (protocol.RemoveResult, error) {
	path := filepath.Join(workspaceRoot, "Cargo.toml")
	src, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return protocol.RemoveResult{}, nil
	}
	if err != nil {
		return protocol.RemoveResult{}, err
	}
	out, changed := DropMember(string(src), params.Project.Root)
	if !changed {
		return protocol.RemoveResult{}, nil
	}
	if !params.DryRun {
		if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
			return protocol.RemoveResult{}, err
		}
	}
	return protocol.RemoveResult{Files: []string{"Cargo.toml"}}, nil
}
