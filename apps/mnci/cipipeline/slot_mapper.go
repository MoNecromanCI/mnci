package cipipeline

import (
	"regexp"
	"strings"
)

var slotPattern = regexp.MustCompile(`(?ms)^([ \t]*)# mnci:slot ([\w-]+)[ \t]*\r?\n(.*?)^[ \t]*# mnci:slot-end ([\w-]+)[ \t]*\r?$`)

// KeepSlots copies the content between each mnci:slot marker pair of the old
// file into the same slot of the newly rendered one. A slot the new file no
// longer has is reported, so a team step is never dropped silently.
func KeepSlots(old, rendered string) (string, []string) {
	bodies := map[string]string{}
	for _, m := range slotPattern.FindAllStringSubmatch(old, -1) {
		if m[2] == m[4] && strings.TrimSpace(m[3]) != "" {
			bodies[m[2]] = m[3]
		}
	}
	used := map[string]bool{}
	out := slotPattern.ReplaceAllStringFunc(rendered, func(block string) string {
		m := slotPattern.FindStringSubmatch(block)
		body, ok := bodies[m[2]]
		if !ok {
			return block
		}
		used[m[2]] = true
		return m[1] + "# mnci:slot " + m[2] + "\n" + body + m[1] + "# mnci:slot-end " + m[2]
	})
	var lost []string
	for name := range bodies {
		if !used[name] {
			lost = append(lost, name)
		}
	}
	return out, lost
}
