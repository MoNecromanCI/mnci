package dependencies

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// plainSpec is one version constraint sync and up understand: ==, >=, ~=,
// <=, < or > and a version. Compound ranges, exclusions (!=), markers, extras
// and URLs are left alone.
var plainSpec = regexp.MustCompile(`^(==|>=|~=|<=|<|>)\s*\d[\w.]*$`)

// sectionOf maps a declaration's section to where it lives in
// pyproject.toml.
var sectionOf = map[string][2]string{
	runtimeKey: {runtimeTable, runtimeKey},
	devKey:     {devTable, devKey},
}

// splitRequirement splits a PEP 508 requirement into its name as written and
// the rest (extras, constraints, markers).
func splitRequirement(req string) (name, rest string) {
	req = strings.TrimSpace(req)
	if loc := nameEnd.FindStringIndex(req); loc != nil {
		return req[:loc[0]], strings.TrimSpace(req[loc[0]:])
	}
	return req, ""
}

// Declared lists the runtime and dev requirements of each project's
// pyproject.toml.
func Declared(workspaceRoot string, params protocol.DeclaredParams) (protocol.DeclaredResult, error) {
	var res protocol.DeclaredResult
	for _, p := range params.Projects {
		file := p.Root + "/pyproject.toml"
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(file)))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return res, err
		}
		for _, section := range []string{runtimeKey, devKey} {
			where := sectionOf[section]
			for _, req := range findList(string(raw), where[0], where[1]).items {
				name, rest := splitRequirement(req)
				res.Declarations = append(res.Declarations, protocol.Declaration{
					Project: p.Name, File: file, Name: RequirementName(name), Spec: rest, Section: section,
					Rewritable: plainSpec.MatchString(rest),
				})
			}
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

// Rewrite changes requirements' constraints in place, keeping each name as
// it is written.
func Rewrite(workspaceRoot string, params protocol.RewriteParams) (protocol.DependencyResult, error) {
	var files []string
	seen := map[string]bool{}
	for _, c := range params.Changes {
		where, ok := sectionOf[c.Declaration.Section]
		if !ok {
			return protocol.DependencyResult{}, fmt.Errorf("%s has no section %q", c.Declaration.File, c.Declaration.Section)
		}
		project := protocol.Project{Name: c.Declaration.Project, Root: strings.TrimSuffix(c.Declaration.File, "/pyproject.toml")}
		err := editPyproject(workspaceRoot, project, func(src string) (string, error) {
			for _, req := range findList(src, where[0], where[1]).items {
				if name, _ := splitRequirement(req); RequirementName(name) == c.Declaration.Name {
					return AddRequirements(src, where[0], where[1], []string{name + c.Spec}), nil
				}
			}
			return "", fmt.Errorf("%s no longer requires %s", c.Declaration.File, c.Declaration.Name)
		})
		if err != nil {
			return protocol.DependencyResult{}, err
		}
		if !seen[c.Declaration.File] {
			seen[c.Declaration.File] = true
			files = append(files, c.Declaration.File)
		}
	}
	return protocol.DependencyResult{Files: files}, nil
}
