package scaffolding

import (
	"errors"
	"regexp"
	"strings"
)

var (
	workspaceKey = regexp.MustCompile(`(?m)^workspace:[ \t]*\n`)
	listItem     = regexp.MustCompile(`^[ \t]+-[ \t]*(.+?)[ \t]*$`)
)

// NewWorkspace is a root pubspec.yaml whose workspace lists dir.
func NewWorkspace(name, dir string) string {
	return "name: " + name + "_workspace\npublish_to: none\n\nenvironment:\n  sdk: " + SDKConstraint + "\n\nworkspace:\n  - " + dir + "\n"
}

// members returns the span of the workspace list's lines and their entries.
func members(src string) (start, end int, entries []string, ok bool) {
	loc := workspaceKey.FindStringIndex(src)
	if loc == nil {
		return 0, 0, nil, false
	}
	start, end = loc[1], loc[1]
	for end < len(src) {
		next := strings.IndexByte(src[end:], '\n')
		line := src[end:]
		if next >= 0 {
			line = src[end : end+next]
		}
		m := listItem.FindStringSubmatch(line)
		if m == nil {
			break
		}
		entries = append(entries, strings.Trim(m[1], `"'`))
		if next < 0 {
			end = len(src)
			break
		}
		end += next + 1
	}
	return start, end, entries, true
}

// AddMember lists dir in the root pubspec's workspace.
func AddMember(src, dir string) (string, error) {
	_, end, entries, ok := members(src)
	if !ok {
		return "", errors.New("the root pubspec.yaml has no workspace: list")
	}
	for _, e := range entries {
		if e == dir {
			return src, nil
		}
	}
	if end == len(src) && !strings.HasSuffix(src, "\n") {
		src += "\n"
		end++
	}
	return src[:end] + "  - " + dir + "\n" + src[end:], nil
}

// DropMember takes dir out of the root pubspec's workspace.
func DropMember(src, dir string) (string, bool) {
	start, end, _, ok := members(src)
	if !ok {
		return src, false
	}
	var kept []string
	changed := false
	for _, line := range strings.SplitAfter(src[start:end], "\n") {
		if m := listItem.FindStringSubmatch(strings.TrimRight(line, "\n")); m != nil && strings.Trim(m[1], `"'`) == dir {
			changed = true
			continue
		}
		kept = append(kept, line)
	}
	return src[:start] + strings.Join(kept, "") + src[end:], changed
}
