package releasing

import "regexp"

var versionField = regexp.MustCompile(`(\.version\s*=\s*)(['"])[^'"]*['"]`)

// SetVersion writes version into the gemspec's version assignment, keeping
// its quotes; a gemspec that computes its version is left alone.
func SetVersion(src, version string) (string, bool) {
	loc := versionField.FindStringSubmatchIndex(src)
	if loc == nil {
		return src, false
	}
	quote := src[loc[4]:loc[5]]
	out := src[:loc[0]] + src[loc[2]:loc[3]] + quote + version + quote + src[loc[1]:]
	return out, out != src
}
