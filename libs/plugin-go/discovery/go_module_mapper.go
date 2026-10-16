package discovery

import (
	"errors"
	"regexp"
)

var (
	modulePattern = regexp.MustCompile(`(?m)^module\s+"?([^\s"]+)"?\s*$`)
	kindPattern   = regexp.MustCompile(`(?m)^//\s*mnci:kind\s+(\w+)\s*$`)
	nativePattern = regexp.MustCompile(`(?m)^//\s*mnci:native\s*$`)
	stackPattern  = regexp.MustCompile(`(?m)^//\s*mnci:stack\s+([\w.@/-]+)\s*$`)
)

// Module is what the plugin reads from a go.mod.
type Module struct {
	Path string
	// Kind is the value of a "// mnci:kind <kind>" comment, if present.
	Kind string
	// Stack is the value of a "// mnci:stack <stack>" comment, if present.
	Stack string
	// Native is set by a "// mnci:native" comment: the module uses cgo, so
	// it is built on each OS rather than cross-compiled.
	Native bool
}

// ParseModule reads the module path, the declared kind and the stack from
// go.mod text.
func ParseModule(src string) (Module, error) {
	m := modulePattern.FindStringSubmatch(src)
	if m == nil {
		return Module{}, errors.New("go.mod has no module line")
	}
	mod := Module{Path: m[1]}
	if k := kindPattern.FindStringSubmatch(src); k != nil {
		mod.Kind = k[1]
	}
	if s := stackPattern.FindStringSubmatch(src); s != nil {
		mod.Stack = s[1]
	}
	mod.Native = nativePattern.MatchString(src)
	return mod, nil
}
