package tui

import (
	"fmt"
	"strings"

	"github.com/MoNecromanCI/mnci/apps/mnci/release"
)

// releaseView shows what the next release would do; releasing itself leaves
// the shell for `mnci release`, which confirms and streams its progress.
type releaseView struct {
	loaded   bool
	plan     release.Plan
	blockers []string
	warnings []string
	err      error
}

type releaseLoaded struct {
	plan     release.Plan
	blockers []string
	warnings []string
	err      error
}

func (v releaseView) view() string {
	if v.err != nil {
		return errorStyle.Render(v.err.Error()) + "\n"
	}
	if !v.loaded {
		return hintStyle.Render("Planning the release…") + "\n"
	}
	var b strings.Builder
	for _, w := range v.warnings {
		b.WriteString(hintStyle.Render("warning: "+w) + "\n")
	}
	if len(v.blockers) > 0 {
		b.WriteString(errorStyle.Render("The release is blocked: its tags would give wrong versions.") + "\n\n")
		for _, m := range v.blockers {
			b.WriteString("  • " + m + "\n")
		}
		b.WriteString(hintStyle.Render("\np mnci release --dry-run (shows the same)") + "\n")
		return b.String()
	}
	if len(v.plan.Items) == 0 {
		b.WriteString("Nothing to release.\n")
	}
	for _, item := range v.plan.Items {
		from := item.Current
		if from == "" {
			from = "new"
		}
		fmt.Fprintf(&b, "  %-24s %s → %s  %s\n", item.Project, from, cursorStyle.Render(item.Next), hintStyle.Render(item.Reason))
	}
	if len(v.plan.Unchanged) > 0 {
		b.WriteString(hintStyle.Render("\nunchanged: "+strings.Join(v.plan.Unchanged, ", ")) + "\n")
	}
	help := "\np preview (mnci release --dry-run)"
	if len(v.plan.Items) > 0 {
		help += " · R release (mnci release: confirms, then streams)"
	}
	b.WriteString(hintStyle.Render(help) + "\n")
	return b.String()
}
