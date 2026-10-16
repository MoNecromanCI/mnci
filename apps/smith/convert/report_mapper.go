package convert

import (
	"fmt"
	"sort"
	"strings"

	"github.com/MoNecromanCI/mnci/apps/mnci/convertplan"
)

var sections = []struct {
	severity convertplan.Severity
	title    string
}{
	{convertplan.Blocker, "Blocking the conversion"},
	{convertplan.Manual, "Needs you"},
	{convertplan.Unsupported, "Not supported by mnci yet"},
	{convertplan.Changed, "Behaves differently"},
	{convertplan.Mapped, "Converted"},
	{convertplan.Dropped, "Dropped (Nx-only)"},
}

// Report renders a plan for a terminal.
func Report(plan *convertplan.Plan, dryRun bool) string {
	var b strings.Builder
	kind := "Nx workspace"
	if plan.Legacy {
		kind = "Nx workspace made by the old mnci"
	}
	verb := "Converting"
	if dryRun {
		verb = "Conversion plan for"
	}
	fmt.Fprintf(&b, "%s %s %q\n\n", verb, kind, plan.Config.Workspace.Name)
	b.WriteString("Projects:\n")
	for _, p := range plan.Projects {
		if !p.Supported {
			fmt.Fprintf(&b, "  %-20s %-28s not supported\n", p.Name, p.Root)
			continue
		}
		fmt.Fprintf(&b, "  %-20s %-28s %s %s\n", p.Name, p.Root, p.Language, p.Kind)
	}
	for _, s := range sections {
		var lines []string
		for _, f := range plan.Findings {
			if f.Severity != s.severity {
				continue
			}
			where := f.Subject
			if f.Project != "" {
				where = f.Project + " " + f.Subject
			}
			lines = append(lines, fmt.Sprintf("  - %s: %s", where, f.Message))
		}
		if len(lines) == 0 {
			continue
		}
		sort.Strings(lines)
		fmt.Fprintf(&b, "\n%s (%d):\n%s\n", s.title, len(lines), strings.Join(lines, "\n"))
	}
	if len(plan.Delete) > 0 {
		fmt.Fprintf(&b, "\nRemoves: %s\n", strings.Join(plan.Delete, ", "))
	}
	return b.String()
}
