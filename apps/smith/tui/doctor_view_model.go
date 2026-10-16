package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/doctor"
)

// doctorView lists doctor's findings, problems first; enter opens the
// selected one's detail and remedy.
type doctorView struct {
	report *doctor.Report
	err    error
	cursor int
	open   bool
}

type doctorLoaded struct {
	report doctor.Report
}

func (v doctorView) update(key tea.KeyMsg) doctorView {
	if v.report == nil || len(v.report.Findings) == 0 {
		return v
	}
	switch key.String() {
	case "up", "k":
		v.cursor = (v.cursor - 1 + len(v.report.Findings)) % len(v.report.Findings)
		v.open = false
	case "down", "j":
		v.cursor = (v.cursor + 1) % len(v.report.Findings)
		v.open = false
	case "enter", " ":
		v.open = !v.open
	}
	return v
}

func (v doctorView) view() string {
	if v.err != nil {
		return errorStyle.Render(v.err.Error()) + "\n"
	}
	if v.report == nil {
		return hintStyle.Render("Checking the workspace…") + "\n"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d passed · %d failed · %d warning(s)\n\n", v.report.Passed, v.report.Failed, v.report.Warnings)
	for i, f := range v.report.Findings {
		mark := "✔"
		switch {
		case !f.OK && f.Warning:
			mark = "!"
		case !f.OK:
			mark = "✖"
		}
		b.WriteString(line(i == v.cursor, mark+" ", f.Check, f.Source))
		if i == v.cursor && v.open && !f.OK {
			if f.Detail != "" {
				b.WriteString("      " + f.Detail + "\n")
			}
			if f.Remedy != "" {
				b.WriteString("      " + cursorStyle.Render("remedy: ") + f.Remedy + "\n")
			}
		}
	}
	b.WriteString(hintStyle.Render("\n↑/↓ move · enter details and remedy · r re-check · f fix (smith doctor --fix)") + "\n")
	return b.String()
}
