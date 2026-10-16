package releasing

import "regexp"

var (
	versionLine = regexp.MustCompile(`(?m)^version:[ \t]*[^\n]*$`)
	nameLine    = regexp.MustCompile(`(?m)^name:[ \t]*[^\n]*\n`)
)

// SetVersion writes version as the pubspec's version, adding the line after
// name: when there is none.
func SetVersion(src, version string) (string, bool) {
	line := "version: " + version
	if loc := versionLine.FindStringIndex(src); loc != nil {
		out := src[:loc[0]] + line + src[loc[1]:]
		return out, out != src
	}
	loc := nameLine.FindStringIndex(src)
	if loc == nil {
		return src, false
	}
	return src[:loc[1]] + line + "\n" + src[loc[1]:], true
}
