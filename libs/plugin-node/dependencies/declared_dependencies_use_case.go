package dependencies

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var sections = []string{"dependencies", "devDependencies", "optionalDependencies", "peerDependencies"}

// plainRange is a spec sync and up understand: an optional operator, then a
// version. Not a tag, URL, git ref, file: or workspace: alias, npm: alias,
// or a compound range.
var plainRange = regexp.MustCompile(`^[\^~>=<]*\d[\w.+-]*$`)

// Declared lists what the root package.json and each project's declare, and
// the versions installed at the root's node_modules.
func Declared(workspaceRoot string, params protocol.DeclaredParams) (protocol.DeclaredResult, error) {
	type manifest struct{ project, file string }
	manifests := []manifest{{"", "package.json"}}
	for _, p := range params.Projects {
		manifests = append(manifests, manifest{p.Name, p.Root + "/package.json"})
	}
	res := protocol.DeclaredResult{Installed: map[string]string{}}
	for _, m := range manifests {
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(m.file)))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return res, err
		}
		var pkg map[string]json.RawMessage
		if err := json.Unmarshal(raw, &pkg); err != nil {
			return res, fmt.Errorf("%s: %w", m.file, err)
		}
		for _, section := range sections {
			var deps map[string]string
			if raw, ok := pkg[section]; !ok || json.Unmarshal(raw, &deps) != nil {
				continue
			}
			for name, spec := range deps {
				peer := section == "peerDependencies"
				res.Declarations = append(res.Declarations, protocol.Declaration{
					Project: m.project, File: m.file, Name: name, Spec: spec, Section: section,
					Peer: peer, Rewritable: !peer && plainRange.MatchString(spec),
				})
				if _, seen := res.Installed[name]; !seen {
					res.Installed[name] = installedVersion(workspaceRoot, name)
				}
			}
		}
	}
	for name, v := range res.Installed {
		if v == "" {
			delete(res.Installed, name)
		}
	}
	sort.SliceStable(res.Declarations, func(i, j int) bool {
		a, b := res.Declarations[i], res.Declarations[j]
		if a.File != b.File {
			return a.File < b.File
		}
		return a.Name < b.Name
	})
	return res, nil
}

func installedVersion(workspaceRoot, name string) string {
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, "node_modules", filepath.FromSlash(name), "package.json"))
	if err != nil {
		return ""
	}
	var pkg struct{ Version string }
	_ = json.Unmarshal(raw, &pkg)
	return pkg.Version
}

// Rewrite changes declared specs in place; every other key, and the order of
// keys, stays.
func Rewrite(workspaceRoot string, params protocol.RewriteParams) (protocol.DependencyResult, error) {
	byFile := map[string][]protocol.Rewrite{}
	var files []string
	for _, c := range params.Changes {
		if _, ok := byFile[c.Declaration.File]; !ok {
			files = append(files, c.Declaration.File)
		}
		byFile[c.Declaration.File] = append(byFile[c.Declaration.File], c)
	}
	sort.Strings(files)
	for _, file := range files {
		path := filepath.Join(workspaceRoot, filepath.FromSlash(file))
		raw, err := os.ReadFile(path)
		if err != nil {
			return protocol.DependencyResult{}, err
		}
		doc, err := jsondoc.ParseDocument(raw)
		if err != nil {
			return protocol.DependencyResult{}, fmt.Errorf("%s: %w", file, err)
		}
		for _, c := range byFile[file] {
			section := doc.Object(c.Declaration.Section, false)
			if section == nil || !section.Has(c.Declaration.Name) {
				return protocol.DependencyResult{}, fmt.Errorf("%s no longer declares %s in %s", file, c.Declaration.Name, c.Declaration.Section)
			}
			section.SetString(c.Declaration.Name, c.Spec)
		}
		if err := os.WriteFile(path, doc.Bytes(), 0o644); err != nil {
			return protocol.DependencyResult{}, err
		}
	}
	return protocol.DependencyResult{Files: files}, nil
}
