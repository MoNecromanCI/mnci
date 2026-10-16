package dependencies

import (
	"regexp"
	"strings"
)

var (
	blockEnd   = regexp.MustCompile(`(?m)^end\s*$`)
	blockParam = regexp.MustCompile(`Gem::Specification\.new\s+do\s*\|(\w+)\|`)
	depCall    = `^[ \t]*\w+\.add_(?:runtime_|development_)?dependency[\s(]+['"]%s['"][^\n]*\n`
)

// AddDependencies declares gems in the gemspec, before its closing end:
// add_dependency, or add_development_dependency when dev. A gem the gemspec
// already declares is left as it is.
func AddDependencies(src string, gems []string, dev bool) string {
	param := "spec"
	if m := blockParam.FindStringSubmatch(src); m != nil {
		param = m[1]
	}
	call := "add_dependency"
	if dev {
		call = "add_development_dependency"
	}
	var lines strings.Builder
	for _, gem := range gems {
		if regexp.MustCompile("(?m)" + strings.Replace(depCall, "%s", regexp.QuoteMeta(gem), 1)).MatchString(src) {
			continue
		}
		lines.WriteString("  " + param + "." + call + " '" + gem + "'\n")
	}
	all := blockEnd.FindAllStringIndex(src, -1)
	if lines.Len() == 0 || len(all) == 0 {
		return src
	}
	at := all[len(all)-1][0]
	return src[:at] + lines.String() + src[at:]
}

// RemoveDependencies takes gems' dependency lines out of the gemspec.
func RemoveDependencies(src string, gems []string) string {
	for _, gem := range gems {
		src = regexp.MustCompile("(?m)"+strings.Replace(depCall, "%s", regexp.QuoteMeta(gem), 1)).ReplaceAllString(src, "")
	}
	return src
}
