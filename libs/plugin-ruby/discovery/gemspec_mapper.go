package discovery

import (
	"regexp"
	"sort"
)

// Gemspec is what smith reads from a .gemspec: the plain assignments and
// dependency calls a gemspec states, not what arbitrary Ruby would compute.
type Gemspec struct {
	Name, Version string
	// Kind and Stack are the smith_kind and smith_stack metadata.
	Kind, Stack string
	// Executables are the gem's executables (in exe/ by default).
	Executables []string
	// Dependencies are the names of every dependency, development ones too.
	Dependencies []string
}

var (
	quoted      = `['"]([^'"]+)['"]`
	nameField   = regexp.MustCompile(`\.name\s*=\s*` + quoted)
	version     = regexp.MustCompile(`\.version\s*=\s*` + quoted)
	metadata    = regexp.MustCompile(`\.metadata\[['"](smith_kind|smith_stack)['"]\]\s*=\s*` + quoted)
	executables = regexp.MustCompile(`\.executables\s*=\s*\[([^\]]*)\]`)
	dependency  = regexp.MustCompile(`\.add_(?:runtime_|development_)?dependency[\s(]+` + quoted)
	item        = regexp.MustCompile(quoted)
)

// ParseGemspec reads a gemspec's name, version, metadata and dependencies.
func ParseGemspec(src string) Gemspec {
	var g Gemspec
	if m := nameField.FindStringSubmatch(src); m != nil {
		g.Name = m[1]
	}
	if m := version.FindStringSubmatch(src); m != nil {
		g.Version = m[1]
	}
	for _, m := range metadata.FindAllStringSubmatch(src, -1) {
		if m[1] == "smith_kind" {
			g.Kind = m[2]
		} else {
			g.Stack = m[2]
		}
	}
	if m := executables.FindStringSubmatch(src); m != nil {
		for _, e := range item.FindAllStringSubmatch(m[1], -1) {
			g.Executables = append(g.Executables, e[1])
		}
	}
	for _, m := range dependency.FindAllStringSubmatch(src, -1) {
		g.Dependencies = append(g.Dependencies, m[1])
	}
	sort.Strings(g.Dependencies)
	return g
}
