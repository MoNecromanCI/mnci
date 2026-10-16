package releasing

import (
	"regexp"
	"strings"
)

var (
	versionElement = regexp.MustCompile(`<Version>[^<]*</Version>`)
	propertyGroup  = regexp.MustCompile(`<PropertyGroup>`)
)

// SetVersion writes version into the project file's first <Version>, adding
// one to the first <PropertyGroup> when there is none.
func SetVersion(src, version string) (string, bool) {
	element := "<Version>" + version + "</Version>"
	if loc := versionElement.FindStringIndex(src); loc != nil {
		out := src[:loc[0]] + element + src[loc[1]:]
		return out, out != src
	}
	loc := propertyGroup.FindStringIndex(src)
	if loc == nil {
		return src, false
	}
	indent := "    "
	if nl := strings.LastIndex(src[:loc[0]], "\n"); nl >= 0 {
		indent = src[nl+1:loc[0]] + "  "
	}
	return src[:loc[1]] + "\n" + indent + element + src[loc[1]:], true
}
