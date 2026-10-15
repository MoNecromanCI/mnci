package discovery

import (
	"errors"
	"regexp"
)

var (
	modulePattern = regexp.MustCompile(`(?m)^module\s+"?([^\s"]+)"?\s*$`)
	kindPattern   = regexp.MustCompile(`(?m)^//\s*mnci:kind\s+(\w+)\s*$`)
)

// Module is what the plugin reads from a go.mod.
type Module struct {
	Path string
	// Kind is the value of a "// mnci:kind <kind>" comment, if present.
	Kind string
}

// ParseModule reads the module path and the declared kind from go.mod text.
func ParseModule(src string) (Module, error) {
	m := modulePattern.FindStringSubmatch(src)
	if m == nil {
		return Module{}, errors.New("go.mod has no module line")
	}
	mod := Module{Path: m[1]}
	if k := kindPattern.FindStringSubmatch(src); k != nil {
		mod.Kind = k[1]
	}
	return mod, nil
}
