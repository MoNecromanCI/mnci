package releasing

import (
	"regexp"
	"strings"
)

var (
	scriptVersion   = regexp.MustCompile(`(?m)^version\s*=\s*["'][^"']*["']`)
	propertyVersion = regexp.MustCompile(`(?m)^version\s*[=:].*$`)
	scriptGroup     = regexp.MustCompile(`(?m)^group\s*=.*$`)
)

// SetScriptVersion writes version as a build script's top-level version,
// adding it after the group (or at the end) when the script sets none.
func SetScriptVersion(src, version string) (string, bool) {
	line := `version = "` + version + `"`
	if loc := scriptVersion.FindStringIndex(src); loc != nil {
		out := src[:loc[0]] + line + src[loc[1]:]
		return out, out != src
	}
	if loc := scriptGroup.FindStringIndex(src); loc != nil {
		return src[:loc[1]] + "\n" + line + src[loc[1]:], true
	}
	if src != "" && !strings.HasSuffix(src, "\n") {
		src += "\n"
	}
	return src + "\n" + line + "\n", true
}

// SetPropertyVersion writes version into gradle.properties when it holds
// the module's version.
func SetPropertyVersion(src, version string) (string, bool) {
	out := propertyVersion.ReplaceAllString(src, "version="+version)
	return out, out != src
}
