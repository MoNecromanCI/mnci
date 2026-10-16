package releasing

import "regexp"

var versionMetadata = regexp.MustCompile(`(?m)^metadata\s+version\s*=\s*'[^']*'`)

// SetVersion writes version as main.bicep's metadata version, adding the
// statement at the top when there is none.
func SetVersion(src, version string) (string, bool) {
	line := "metadata version = '" + version + "'"
	if loc := versionMetadata.FindStringIndex(src); loc != nil {
		out := src[:loc[0]] + line + src[loc[1]:]
		return out, out != src
	}
	return line + "\n" + src, true
}
