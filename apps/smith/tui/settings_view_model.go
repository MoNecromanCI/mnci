package tui

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/configedit"
)

// setting is one row of the settings view: a field of smith.toml and its value
// as written (set is false when it has its default).
type setting struct {
	field configedit.Field
	value string
	set   bool
}

type settingsLoaded struct {
	settings []setting
	err      error
}

// settingsView lists smith.toml's settings. Enter edits one (a boolean
// toggles), checked against the schema before it leaves the shell as `smith
// config set`; u returns one to its default with `smith config unset`.
type settingsView struct {
	loaded   bool
	settings []setting
	err      error
	cursor   int
	editing  bool
	input    textinput.Model
	problem  string
}

var quoted = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`)

// args are a value as written, as the arguments `smith config set` takes.
func args(f configedit.Field, value string) []string {
	switch f.Type {
	case "array":
		var out []string
		for _, m := range quoted.FindAllStringSubmatch(value, -1) {
			out = append(out, m[1])
		}
		return out
	case "string":
		if s, err := strconv.Unquote(value); err == nil {
			return []string{s}
		}
	}
	if value == "" {
		return nil
	}
	return []string{value}
}

// typed splits what was typed into arguments: a list takes words, anything
// else is one value.
func typed(f configedit.Field, text string) []string {
	if f.Type == "array" {
		return strings.Fields(text)
	}
	return []string{strings.TrimSpace(text)}
}

// update handles a key; argv is set when the user confirmed a change.
func (v settingsView) update(key tea.KeyMsg) (settingsView, []string) {
	if !v.loaded || len(v.settings) == 0 {
		return v, nil
	}
	s := v.settings[v.cursor]
	if v.editing {
		switch key.Type {
		case tea.KeyEsc:
			v.editing, v.problem = false, ""
			return v, nil
		case tea.KeyEnter:
			values := typed(s.field, v.input.Value())
			if _, err := s.field.Encode(values); err != nil {
				v.problem = err.Error()
				return v, nil
			}
			return v, append([]string{"config", "set", s.field.Path}, values...)
		}
		v.input, _ = v.input.Update(key)
		return v, nil
	}
	switch key.String() {
	case "up", "k":
		v.cursor = (v.cursor - 1 + len(v.settings)) % len(v.settings)
	case "down", "j":
		v.cursor = (v.cursor + 1) % len(v.settings)
	case "u":
		if s.set {
			return v, []string{"config", "unset", s.field.Path}
		}
		v.problem = s.field.Path + " already has its default"
	case "enter":
		if s.field.Type == "boolean" {
			return v, []string{"config", "set", s.field.Path, strconv.FormatBool(s.value != "true")}
		}
		v.input = textinput.New()
		v.input.SetValue(strings.Join(args(s.field, s.value), " "))
		v.input.Focus()
		v.editing, v.problem = true, ""
	}
	return v, nil
}

func (v settingsView) view() string {
	if v.err != nil {
		return errorStyle.Render(v.err.Error()) + "\n"
	}
	if !v.loaded {
		return hintStyle.Render("Reading smith.toml…") + "\n"
	}
	var b strings.Builder
	for i, s := range v.settings {
		value := s.value
		if !s.set {
			value = hintStyle.Render("(default)")
		}
		b.WriteString(line(i == v.cursor, "", s.field.Path, value))
	}
	s := v.settings[v.cursor]
	b.WriteString("\n" + hintStyle.Render(s.field.Description) + "\n")
	if v.editing {
		hint := "one value"
		if s.field.Type == "array" {
			hint = "values separated by spaces"
			if len(s.field.Enum) > 0 {
				hint += " (" + strings.Join(s.field.Enum, ", ") + ")"
			}
		}
		b.WriteString("\n" + titleStyle.Render(s.field.Path+" = ") + v.input.View() + "\n")
		b.WriteString(hintStyle.Render(hint+" · enter save (smith config set) · esc cancel") + "\n")
	} else {
		b.WriteString(hintStyle.Render("\n↑/↓ move · enter edit (a yes/no setting flips) · u back to the default") + "\n")
	}
	if v.problem != "" {
		b.WriteString(errorStyle.Render(v.problem) + "\n")
	}
	return b.String()
}
