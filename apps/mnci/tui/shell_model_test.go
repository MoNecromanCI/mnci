package tui

import (
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/apps/mnci/clicontract"
	"github.com/MoNecromanCI/mnci/apps/mnci/configedit"
	"github.com/MoNecromanCI/mnci/apps/mnci/doctor"
	"github.com/MoNecromanCI/mnci/apps/mnci/release"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
	tea "github.com/charmbracelet/bubbletea"
)

var commands = []clicontract.CommandDescription{
	{Name: "affected", Description: "List projects a change touches, before a release"},
	{Name: "doctor", Description: "Check the workspace"},
	{Name: "release", Description: "Version, tag and publish", Options: []clicontract.OptionDescription{{Name: "dry-run"}}},
}

// press sends keys, running each command the shell returns the way Bubble Tea
// would, so a view that loads gets its result.
func press(m tea.Model, ks ...string) tea.Model {
	for _, k := range ks {
		var cmd tea.Cmd
		m, cmd = m.Update(keyMsg(k))
		if cmd != nil {
			if msg := cmd(); msg != nil {
				if _, quit := msg.(tea.QuitMsg); !quit {
					m, _ = m.Update(msg)
				}
			}
		}
	}
	return m
}

func sources(report doctor.Report, loaded releaseLoaded) *Sources {
	return &Sources{
		Doctor:  func() (doctor.Report, error) { return report, nil },
		Release: func() releaseLoaded { return loaded },
		Settings: func() settingsLoaded {
			fields, _ := configedit.Fields()
			var out settingsLoaded
			for _, f := range fields {
				switch f.Path {
				case "ci.devcontainer":
					out.settings = append(out.settings, setting{field: f, value: "true", set: true})
				case "ci.providers":
					out.settings = append(out.settings, setting{field: f, value: `["github"]`, set: true})
				case "release.branch":
					out.settings = append(out.settings, setting{field: f})
				}
			}
			return out
		},
	}
}

func TestTheMenuTabBuildsACommand(t *testing.T) {
	sh := press(NewShell(menu(), commands, nil), "down", "enter", "down", "down", "enter", "down", "enter", "s", "d", "k", "enter")
	if got := strings.Join(sh.(Shell).Argv, " "); got != "add package sdk --lang node" {
		t.Fatalf("argv %q", got)
	}
}

func TestOutsideAWorkspaceThereIsNoDoctorOrRelease(t *testing.T) {
	view := NewShell(menu(), commands, nil).View()
	if !strings.Contains(view, "Commands") || strings.Contains(view, "Doctor") {
		t.Fatalf("tabs:\n%s", view)
	}
}

func TestThePaletteFindsAndRunsACommand(t *testing.T) {
	sh := press(NewShell(menu(), commands, nil), "tab", "r", "e", "l")
	if view := sh.View(); !strings.Contains(view, "› release") {
		t.Fatalf("release should lead the matches:\n%s", view)
	}
	sh = press(sh, "enter")
	if view := sh.View(); !strings.Contains(view, "mnci release") || !strings.Contains(view, "[--dry-run]") {
		t.Fatalf("picked:\n%s", view)
	}
	sh = press(sh, "-", "-", "d", "r", "y", "-", "r", "u", "n", "enter")
	if got := strings.Join(sh.(Shell).Argv, " "); got != "release --dry-run" {
		t.Fatalf("argv %q", got)
	}
}

func TestTheDoctorViewOpensARemedy(t *testing.T) {
	report := doctor.Report{Passed: 1, Failed: 1, Findings: []doctor.Finding{
		{Finding: protocol.Finding{Check: "go.work lists every module", Detail: "libs/x is missing", Remedy: "go work use ./libs/x"}, Source: "go"},
		{Finding: protocol.Finding{Check: "git is installed", OK: true}, Source: "mnci"},
	}}
	sh := press(NewShell(menu(), commands, sources(report, releaseLoaded{})), "tab", "tab")
	view := sh.View()
	if !strings.Contains(view, "1 passed · 1 failed") || !strings.Contains(view, "✖ go.work lists every module") || strings.Contains(view, "remedy:") {
		t.Fatalf("doctor:\n%s", view)
	}
	if view := press(sh, "enter").View(); !strings.Contains(view, "remedy: go work use ./libs/x") || !strings.Contains(view, "libs/x is missing") {
		t.Fatalf("opened:\n%s", view)
	}
}

func TestTheReleaseViewShowsThePlanAndHandsOver(t *testing.T) {
	plan := release.Plan{Items: []release.Item{{Project: "web", Current: "1.0.0", Next: "1.1.0", Reason: "feat: dark mode"}}, Unchanged: []string{"sdk"}}
	sh := press(NewShell(menu(), commands, sources(doctor.Report{}, releaseLoaded{plan: plan})), "tab", "tab", "tab")
	view := sh.View()
	if !strings.Contains(view, "web") || !strings.Contains(view, "1.0.0 → ") || !strings.Contains(view, "1.1.0") || !strings.Contains(view, "unchanged: sdk") {
		t.Fatalf("release:\n%s", view)
	}
	if got := strings.Join(press(sh, "p").(Shell).Argv, " "); got != "release --dry-run" {
		t.Fatalf("preview argv %q", got)
	}
	if got := strings.Join(press(sh, "R").(Shell).Argv, " "); got != "release" {
		t.Fatalf("release argv %q", got)
	}

	blocked := press(NewShell(menu(), commands, sources(doctor.Report{}, releaseLoaded{plan: plan, blockers: []string{"web is tagged 1.0.0 but npm already has 1.2.0"}})), "tab", "tab", "tab")
	if view := blocked.View(); !strings.Contains(view, "blocked") || !strings.Contains(view, "npm already has 1.2.0") {
		t.Fatalf("blocked:\n%s", view)
	}
	if argv := press(blocked, "R").(Shell).Argv; argv != nil {
		t.Fatalf("a blocked release must not start: %v", argv)
	}
}

func TestTheSettingsViewEditsThroughMnciConfig(t *testing.T) {
	sh := press(NewShell(menu(), commands, sources(doctor.Report{}, releaseLoaded{})), "shift+tab")
	view := sh.View()
	if !strings.Contains(view, "ci.devcontainer") || !strings.Contains(view, `["github"]`) || !strings.Contains(view, "(default)") {
		t.Fatalf("settings:\n%s", view)
	}
	// Rows are in schema order: devcontainer, providers, then release.branch.
	if got := strings.Join(press(sh, "enter").(Shell).Argv, " "); got != "config set ci.devcontainer false" {
		t.Fatalf("a boolean flips: %q", got)
	}
	editing := press(sh, "down", "enter")
	if view := editing.View(); !strings.Contains(view, "ci.providers = ") || !strings.Contains(view, "github") {
		t.Fatalf("editing:\n%s", view)
	}
	rejected := press(editing, " ", "g", "i", "t", "l", "a", "b", "enter")
	if rejected.(Shell).Argv != nil || !strings.Contains(rejected.View(), `not "gitlab"`) {
		t.Fatalf("the schema rejects gitlab:\n%s", rejected.View())
	}
	accepted := press(editing, " ", "a", "z", "u", "r", "e", "enter")
	if got := strings.Join(accepted.(Shell).Argv, " "); got != "config set ci.providers github azure" {
		t.Fatalf("argv %q", got)
	}
	if got := strings.Join(press(sh, "down", "down", "u").(Shell).Argv, " "); got != "" || !strings.Contains(press(sh, "down", "down", "u").View(), "already has its default") {
		t.Fatalf("unsetting a default: %q", got)
	}
	if got := strings.Join(press(sh, "u").(Shell).Argv, " "); got != "config unset ci.devcontainer" {
		t.Fatalf("unset argv %q", got)
	}
}
