package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func keys(m tea.Model, ks ...string) tea.Model {
	for _, k := range ks {
		var msg tea.KeyMsg
		switch k {
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "space":
			msg = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		m, _ = m.Update(msg)
	}
	return m
}

func menu() Model {
	return NewModel("mnci test", Actions([]string{"go", "node"}, []string{"go", "node", "python"}, []string{"go", "node"})[1:])
}

func TestAddAProject(t *testing.T) {
	// Add a project (2nd entry) → package (3rd) → node (2nd) → name.
	m := keys(menu(), "down", "enter", "down", "down", "enter", "down", "enter", "s", "d", "k", "enter")
	got := strings.Join(m.(Model).Argv, " ")
	if got != "add package sdk --lang node" {
		t.Fatalf("argv %q", got)
	}
}

func TestRunTasksDefaultsToEveryTaskAndToggles(t *testing.T) {
	// Run tasks → untoggle lint → confirm → affected.
	m := keys(menu(), "enter", "space", "enter", "down", "enter")
	if got := strings.Join(m.(Model).Argv, " "); got != "run typecheck test build --affected" {
		t.Fatalf("argv %q", got)
	}
}

func TestPluginsStartFromTheirCurrentState(t *testing.T) {
	// Manage plugins (3rd) → python is off; toggle it on.
	m := keys(menu(), "down", "down", "enter")
	view := m.View()
	if !strings.Contains(view, "[x] go") || !strings.Contains(view, "[ ] python") {
		t.Fatalf("view:\n%s", view)
	}
	m = keys(m, "down", "down", "space", "enter")
	if got := strings.Join(m.(Model).Argv, " "); got != "plugin set go node python" {
		t.Fatalf("argv %q", got)
	}
}

func TestValidationAndBack(t *testing.T) {
	m := keys(menu(), "down", "enter", "enter", "enter", "enter")
	if !strings.Contains(m.View(), "type a value") {
		t.Fatalf("an empty name must be refused:\n%s", m.View())
	}
	m = keys(m, "esc")
	if m.(Model).action != nil || len(m.(Model).Argv) != 0 {
		t.Fatal("esc returns to the menu")
	}
	m = keys(m, "q")
	if len(m.(Model).Argv) != 0 {
		t.Fatal("quitting runs nothing")
	}
}
