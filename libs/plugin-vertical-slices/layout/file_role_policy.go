package layout

import (
	"regexp"
	"strings"
)

var (
	platformSuffix = `(?:_(?:` + strings.Join(goPlatforms, "|") + `))*`
	goRoleFile     = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*_(?:` + strings.Join(Roles, "|") + `)` + platformSuffix + `(?:_test)?\.go$`)
	goRootFile     = regexp.MustCompile(`^(?:doc|main)` + platformSuffix + `(?:_test)?\.go$`)
	pyRoleFile     = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*_(?:` + strings.Join(Roles, "|") + `)\.py$`)
	pyTestFile     = regexp.MustCompile(`^test_[a-z0-9_]+\.py$`)
	snakeCase      = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`)
)

// GoFile says why a Go file breaks the rules, or "" when it does not. A
// module's root holds only doc.go and main.go; every other file is
// <snake>_<role>.go.
func GoFile(name string, atRoot bool) string {
	if atRoot {
		if goRootFile.MatchString(name) {
			return ""
		}
		return "only doc.go and main.go sit at the root of a module; move this into a slice folder"
	}
	if name == "doc.go" || goRoleFile.MatchString(name) {
		return ""
	}
	return "a Go file is <snake>_<role>.go (roles: " + strings.Join(Roles, ", ") + ")"
}

// PythonFile says why a Python module breaks the rules, or "" when it does
// not. A package's top folder holds only __init__.py and __main__.py; a slice
// module is <snake>_<role>.py; a test is test_<name>.py.
func PythonFile(name string, atPackageRoot bool) string {
	switch {
	case name == "__init__.py", name == "conftest.py", pyTestFile.MatchString(name):
		return ""
	case name == "__main__.py" && atPackageRoot:
		return ""
	case atPackageRoot:
		return "only __init__.py and __main__.py sit at the top of a package; move this into a slice package"
	case pyRoleFile.MatchString(name):
		return ""
	}
	return "a Python module is <snake>_<role>.py (roles: " + strings.Join(Roles, ", ") + ")"
}

// production reports whether a file counts toward a slice's size: code, not
// a test or a package marker.
func production(name string) bool {
	switch {
	case strings.HasSuffix(name, ".go"):
		return name != "doc.go" && !strings.HasSuffix(name, "_test.go")
	case strings.HasSuffix(name, ".py"):
		return name != "__init__.py" && name != "conftest.py" && !pyTestFile.MatchString(name)
	}
	return false
}
