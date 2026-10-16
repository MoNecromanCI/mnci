package projectkind

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Resolve returns the declared kind when it is valid; otherwise it falls back
// to the directory convention (apps/ → app, libs/ → lib, packages/ → package),
// and finally to fallback.
func Resolve(declared string, root string, fallback protocol.ProjectKind) protocol.ProjectKind {
	switch k := protocol.ProjectKind(strings.TrimSpace(declared)); k {
	case protocol.ProjectApp, protocol.ProjectLib, protocol.ProjectPackage:
		return k
	}
	top, _, _ := strings.Cut(root, "/")
	switch top {
	case "apps":
		return protocol.ProjectApp
	case "libs":
		return protocol.ProjectLib
	case "packages":
		return protocol.ProjectPackage
	}
	return fallback
}

// Directory is the conventional parent directory of a kind.
func Directory(kind protocol.ProjectKind) string {
	switch kind {
	case protocol.ProjectApp:
		return "apps"
	case protocol.ProjectLib:
		return "libs"
	default:
		return "packages"
	}
}
