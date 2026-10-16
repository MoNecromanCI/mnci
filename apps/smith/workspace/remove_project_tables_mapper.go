package workspace

import (
	"regexp"
	"strings"
)

var tableHeader = regexp.MustCompile(`^\s*\[\[?\s*([^\]]+?)\s*\]\]?\s*(#.*)?$`)

// RemoveProjectTables deletes [projects.<name>] and every table under it
// ([projects.<name>.tasks.build], ...) from smith.toml text, keeping all other
// lines exactly. It reports whether anything was removed.
func RemoveProjectTables(src, name string) (string, bool) {
	lines := strings.SplitAfter(src, "\n")
	var out []string
	skipping, removed := false, false
	for _, line := range lines {
		if m := tableHeader.FindStringSubmatch(strings.TrimRight(line, "\r\n")); m != nil {
			skipping = projectTable(m[1], name)
			removed = removed || skipping
		}
		if !skipping {
			out = append(out, line)
		}
	}
	text := strings.Join(out, "")
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}
	return text, removed
}

// projectTable reports whether a table key is projects.<name> or below it.
func projectTable(key, name string) bool {
	rest, ok := strings.CutPrefix(key, "projects.")
	if !ok {
		return false
	}
	for _, form := range []string{name, `"` + name + `"`} {
		if rest == form || strings.HasPrefix(rest, form+".") {
			return true
		}
	}
	return false
}
