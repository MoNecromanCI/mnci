package releasing

import (
	"regexp"
	"strings"
)

var (
	packageHeader = regexp.MustCompile(`(?m)^\[package\][ \t]*(#.*)?\r?\n?`)
	versionLine   = regexp.MustCompile(`(?m)^([ \t]*version[ \t]*=[ \t]*)"[^"]*"`)
	inlineVersion = regexp.MustCompile(`(version[ \t]*=[ \t]*)"[^"]*"`)
	tableHeader   = regexp.MustCompile(`(?m)^\[`)
)

// SetPackageVersion writes version into [package]; changed is false when
// [package] has no literal version (version.workspace = true shares the
// workspace's).
func SetPackageVersion(src, version string) (string, bool) {
	h := packageHeader.FindStringIndex(src)
	if h == nil {
		return src, false
	}
	end := len(src)
	if next := tableHeader.FindStringIndex(src[h[1]:]); next != nil {
		end = h[1] + next[0]
	}
	body := src[h[1]:end]
	loc := versionLine.FindStringSubmatchIndex(body)
	if loc == nil {
		return src, false
	}
	updated := body[:loc[0]] + body[loc[2]:loc[3]] + `"` + version + `"` + body[loc[1]:]
	return src[:h[1]] + updated + src[end:], updated != body
}

// SetDependencyVersions updates the version requirement of inline path
// dependencies on the crates named, keeping the operator the line has
// (a plain version is a caret requirement for cargo).
func SetDependencyVersions(src string, versions map[string]string) (string, bool) {
	changed := false
	lines := strings.SplitAfter(src, "\n")
	for i, line := range lines {
		key, rest, ok := strings.Cut(line, "=")
		if !ok || !strings.Contains(rest, "{") || !strings.Contains(rest, "path") || !inlineVersion.MatchString(rest) {
			continue
		}
		name := strings.TrimSpace(key)
		if m := regexp.MustCompile(`package[ \t]*=[ \t]*"([^"]+)"`).FindStringSubmatch(rest); m != nil {
			name = m[1]
		}
		v, ok := versions[name]
		if !ok {
			continue
		}
		replaced := inlineVersion.ReplaceAllString(rest, `${1}"`+v+`"`)
		if replaced != rest {
			lines[i] = key + "=" + replaced
			changed = true
		}
	}
	return strings.Join(lines, ""), changed
}
