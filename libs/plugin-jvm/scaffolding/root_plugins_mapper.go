package scaffolding

import (
	"regexp"
	"strings"
)

var pluginsBlock = regexp.MustCompile(`(?m)^plugins\s*\{\s*$`)

// AddRootPlugin adds line to the root script's plugins block (pinning a
// plugin's version for the modules), or a block at the end when there is
// none. A plugin the block already pins is left as it is.
func AddRootPlugin(src, line string) (string, bool) {
	id, _, _ := strings.Cut(line, " version ")
	if strings.Contains(src, id) {
		return src, false
	}
	loc := pluginsBlock.FindStringIndex(src)
	if loc == nil {
		if src != "" && !strings.HasSuffix(src, "\n") {
			src += "\n"
		}
		return src + "\nplugins {\n    " + line + "\n}\n", true
	}
	end := strings.Index(src[loc[1]:], "\n}")
	if end < 0 {
		return src, false
	}
	at := loc[1] + end
	return src[:at] + "\n    " + line + src[at:], true
}
