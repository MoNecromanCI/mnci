package dependencies

import (
	"regexp"
	"strings"
)

var (
	nameEnd   = regexp.MustCompile(`[\s<>=!~;\[(@]`)
	separator = regexp.MustCompile(`[-_.]+`)
	element   = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"|'([^']*)'`)
)

// RequirementName is a PEP 508 requirement's distribution name, normalised
// as PEP 503 compares them: "Requests_OAuthlib>=1" is "requests-oauthlib".
func RequirementName(spec string) string {
	spec = strings.TrimSpace(spec)
	if loc := nameEnd.FindStringIndex(spec); loc != nil {
		spec = spec[:loc[0]]
	}
	return strings.ToLower(separator.ReplaceAllString(spec, "-"))
}

// list is a TOML array of strings at key in [table]: where it sits and what
// it holds. tableAt is the end of the table's header line, -1 when the file
// has no such table; start is -1 when the table has no such key.
type list struct {
	tableAt, start, end int
	items               []string
	multiline           bool
}

func findList(src, table, key string) list {
	header := regexp.MustCompile(`(?m)^[ \t]*\[` + regexp.QuoteMeta(table) + `\][ \t]*(#.*)?\r?\n?`)
	h := header.FindStringIndex(src)
	if h == nil {
		return list{tableAt: -1, start: -1}
	}
	bodyEnd := len(src)
	if next := regexp.MustCompile(`(?m)^[ \t]*\[`).FindStringIndex(src[h[1]:]); next != nil {
		bodyEnd = h[1] + next[0]
	}
	assign := regexp.MustCompile(`(?m)^[ \t]*` + regexp.QuoteMeta(key) + `[ \t]*=[ \t]*\[`).FindStringIndex(src[h[1]:bodyEnd])
	if assign == nil {
		return list{tableAt: h[1], start: -1}
	}
	l := list{tableAt: h[1], start: h[1] + assign[1] - 1}
	depth, quote := 0, byte(0)
	for i := l.start; i < len(src); i++ {
		c := src[i]
		switch {
		case quote != 0:
			if c == '\\' && quote == '"' {
				i++
			} else if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '#':
			for i < len(src) && src[i] != '\n' {
				i++
			}
		case c == '[':
			depth++
		case c == ']':
			depth--
			if depth == 0 {
				l.end = i + 1
				inner := src[l.start:l.end]
				l.multiline = strings.Contains(inner, "\n")
				for _, m := range element.FindAllStringSubmatch(inner, -1) {
					l.items = append(l.items, m[1]+m[2])
				}
				return l
			}
		}
	}
	l.start = -1
	return l
}

func render(items []string, multiline bool) string {
	if len(items) == 0 {
		return "[]"
	}
	quoted := make([]string, len(items))
	for i, it := range items {
		quoted[i] = `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(it) + `"`
	}
	if multiline {
		return "[\n    " + strings.Join(quoted, ",\n    ") + ",\n]"
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// AddRequirements puts specs into the array at key in [table], replacing a
// requirement of the same name, creating the key or the table when missing.
// The rest of the file is left as it is.
func AddRequirements(src, table, key string, specs []string) string {
	l := findList(src, table, key)
	items := l.items
	for _, spec := range specs {
		replaced := false
		for i, it := range items {
			if RequirementName(it) == RequirementName(spec) {
				items[i], replaced = spec, true
			}
		}
		if !replaced {
			items = append(items, spec)
		}
	}
	switch {
	case l.tableAt < 0:
		if src != "" && !strings.HasSuffix(src, "\n") {
			src += "\n"
		}
		return src + "\n[" + table + "]\n" + key + " = " + render(items, false) + "\n"
	case l.start < 0:
		return src[:l.tableAt] + key + " = " + render(items, false) + "\n" + src[l.tableAt:]
	}
	return src[:l.start] + render(items, l.multiline) + src[l.end:]
}

// RemoveRequirements drops the requirements named from the array at key in
// [table]; removed lists the names it found.
func RemoveRequirements(src, table, key string, names []string) (out string, removed []string) {
	l := findList(src, table, key)
	if l.start < 0 {
		return src, nil
	}
	var kept []string
	for _, it := range l.items {
		drop := false
		for _, n := range names {
			if RequirementName(it) == RequirementName(n) {
				drop = true
				removed = append(removed, RequirementName(n))
			}
		}
		if !drop {
			kept = append(kept, it)
		}
	}
	if len(removed) == 0 {
		return src, nil
	}
	return src[:l.start] + render(kept, l.multiline && len(kept) > 0) + src[l.end:], removed
}
