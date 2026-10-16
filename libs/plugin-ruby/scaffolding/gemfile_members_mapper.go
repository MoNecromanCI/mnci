package scaffolding

import (
	"regexp"
	"strings"
)

var gemspecLine = regexp.MustCompile(`(?m)^gemspec\s+path:\s*['"]([^'"]+)['"][^\n]*\n`)

// firstGem is the first gem or group line: a new gemspec goes before it.
var firstGem = regexp.MustCompile(`(?m)^(gem|group)\s`)

// AddGemspec lists dir in the Gemfile with gemspec path:, after the gems it
// already lists (or before its first gem or group when it lists none).
func AddGemspec(src, dir string) string {
	line := "gemspec path: '" + dir + "'\n"
	all := gemspecLine.FindAllStringSubmatchIndex(src, -1)
	for _, m := range all {
		if src[m[2]:m[3]] == dir {
			return src
		}
	}
	if len(all) > 0 {
		at := all[len(all)-1][1]
		return src[:at] + line + src[at:]
	}
	if loc := firstGem.FindStringIndex(src); loc != nil {
		return src[:loc[0]] + line + "\n" + src[loc[0]:]
	}
	if src != "" && !strings.HasSuffix(src, "\n") {
		src += "\n"
	}
	return src + line
}

// DropGemspec takes dir's gemspec line out of the Gemfile.
func DropGemspec(src, dir string) (string, bool) {
	out := gemspecLine.ReplaceAllStringFunc(src, func(line string) string {
		if gemspecLine.FindStringSubmatch(line)[1] == dir {
			return ""
		}
		return line
	})
	return out, out != src
}
