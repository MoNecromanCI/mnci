package releasing

import (
	"errors"
	"regexp"
	"strings"
)

var (
	tableHeader = regexp.MustCompile(`^\s*\[\[?\s*([^\]]+?)\s*\]\]?\s*(#.*)?$`)
	versionLine = regexp.MustCompile(`^(\s*version\s*=\s*)(["'])[^"']*(["'])(.*)$`)
)

// SetVersion rewrites `version = "..."` inside the [project] table and nothing
// else, keeping the line's quoting, spacing, comments and line endings.
func SetVersion(src, version string) (string, error) {
	lines := strings.SplitAfter(src, "\n")
	inProject := false
	for i, line := range lines {
		body := strings.TrimRight(line, "\r\n")
		if m := tableHeader.FindStringSubmatch(body); m != nil {
			inProject = m[1] == "project"
			continue
		}
		if !inProject {
			continue
		}
		if m := versionLine.FindStringSubmatch(body); m != nil {
			lines[i] = m[1] + m[2] + version + m[3] + m[4] + line[len(body):]
			return strings.Join(lines, ""), nil
		}
	}
	return "", errors.New(`pyproject.toml has no version in [project]`)
}
