package releasing

import "regexp"

var (
	versionField = regexp.MustCompile(`("version"\s*:\s*)"[^"]*"`)
	nameField    = regexp.MustCompile(`(?m)^([ \t]*)"name"\s*:\s*"[^"]*",[ \t]*\n`)
)

// SetVersion writes version as the composer.json's version, adding the
// field after name when there is none.
func SetVersion(src, version string) (string, bool) {
	if loc := versionField.FindStringSubmatchIndex(src); loc != nil {
		out := src[:loc[0]] + src[loc[2]:loc[3]] + `"` + version + `"` + src[loc[1]:]
		return out, out != src
	}
	loc := nameField.FindStringSubmatchIndex(src)
	if loc == nil {
		return src, false
	}
	indent := src[loc[2]:loc[3]]
	return src[:loc[1]] + indent + `"version": "` + version + `",` + "\n" + src[loc[1]:], true
}
