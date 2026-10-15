package layout

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Check walks a project and reports what breaks the vertical-slice rules.
// language is "go" or "python"; dir is the project root (where go.mod or
// pyproject.toml is).
func Check(dir, language string) (Report, error) {
	var r Report
	switch language {
	case "go", "python":
	default:
		return r, fmt.Errorf("the vertical-slices check knows go and python, not %q", language)
	}
	err := walk(&r, dir, "", language, false)
	return r, err
}

func walk(r *Report, root, rel, language string, inPackage bool) error {
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return err
	}
	shown := func(name string) string { return path.Join(rel, name) }
	isPackage := language == "python" && has(entries, "__init__.py")
	packageRoot := isPackage && !inPackage
	files := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			if strings.HasPrefix(name, ".") || skippedFolders[name] || strings.HasSuffix(name, ".egg-info") {
				continue
			}
			if ForbiddenFolders[name] {
				r.problem(shown(name)+"/", fmt.Sprintf("%q is a junk drawer; name the folder for the outcome it serves", name))
				continue
			}
			if language == "python" && isPackage && !snakeCase.MatchString(name) {
				r.problem(shown(name)+"/", "a Python package is snake_case: its name is an import path")
			}
			if err := walk(r, root, shown(name), language, inPackage || isPackage); err != nil {
				return err
			}
			continue
		}
		var why string
		switch {
		case language == "go" && strings.HasSuffix(name, ".go"):
			why = GoFile(name, rel == "")
		case language == "python" && strings.HasSuffix(name, ".py") && (isPackage || rel != ""):
			why = PythonFile(name, packageRoot)
		default:
			continue
		}
		if why != "" {
			r.problem(shown(name), why)
		}
		if production(name) {
			files++
		}
	}
	if files > SliceSize {
		r.warn(strings.TrimPrefix(rel+"/", "/"), fmt.Sprintf("%d production files (more than %d): it usually holds two outcomes, so split it into two slices or write the exception down", files, SliceSize))
	}
	return nil
}

func has(entries []os.DirEntry, name string) bool {
	for _, e := range entries {
		if e.Name() == name && !e.IsDir() {
			return true
		}
	}
	return false
}
