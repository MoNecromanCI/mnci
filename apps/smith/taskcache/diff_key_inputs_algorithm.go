package taskcache

import (
	"fmt"
	"sort"
	"strings"
)

// Difference is one way a task's key inputs changed.
type Difference struct {
	// What changed: "file", "command", "env", "upstream", "dependency",
	// "tool" or "platform".
	What string `json:"what"`
	// Name is the file, variable, project or task concerned, when there is one.
	Name string `json:"name,omitempty"`
	// Change is "added", "removed" or "changed", or the old → new value.
	Change string `json:"change"`
}

func (d Difference) String() string {
	if d.Name == "" {
		return fmt.Sprintf("%-10s %s", d.What, d.Change)
	}
	return fmt.Sprintf("%-10s %s %s", d.What, d.Name, d.Change)
}

// Diff lists what differs between the inputs of an earlier key and of the
// current one, in a stable order: what was asked to run, then files, then
// what the task depends on.
func Diff(before, now KeyInputs) []Difference {
	var out []Difference
	if before.Tool != now.Tool {
		out = append(out, Difference{What: "tool", Change: before.Tool + " → " + now.Tool})
	}
	if before.Platform != now.Platform {
		out = append(out, Difference{What: "platform", Change: before.Platform + " → " + now.Platform})
	}
	if strings.Join(before.Command, "\x00") != strings.Join(now.Command, "\x00") {
		out = append(out, Difference{What: "command", Change: strings.Join(before.Command, " ") + " → " + strings.Join(now.Command, " ")})
	}
	out = append(out, diffMaps("env", before.Env, now.Env)...)
	out = append(out, diffMaps("file", fileMap(before.Files), fileMap(now.Files))...)
	out = append(out, diffMaps("upstream", before.Upstream, now.Upstream)...)
	out = append(out, diffMaps("dependency", before.Deps, now.Deps)...)
	return out
}

func fileMap(files []FileDigest) map[string]string {
	m := make(map[string]string, len(files))
	for _, f := range files {
		m[f.Path] = f.Digest
	}
	return m
}

func diffMaps(what string, before, now map[string]string) []Difference {
	names := map[string]bool{}
	for k := range before {
		names[k] = true
	}
	for k := range now {
		names[k] = true
	}
	sorted := make([]string, 0, len(names))
	for k := range names {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)
	var out []Difference
	for _, k := range sorted {
		old, had := before[k]
		cur, has := now[k]
		switch {
		case !had:
			out = append(out, Difference{What: what, Name: k, Change: "added"})
		case !has:
			out = append(out, Difference{What: what, Name: k, Change: "removed"})
		case old != cur:
			out = append(out, Difference{What: what, Name: k, Change: "changed"})
		}
	}
	return out
}
