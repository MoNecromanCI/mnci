package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
)

// palette finds any smith command by typing part of its name or description;
// enter takes it, then the rest of the command line can be typed.
type palette struct {
	commands []clicontract.CommandDescription
	filter   textinput.Model
	cursor   int
	// chosen is set once a command is picked; filter then holds its
	// arguments.
	chosen *clicontract.CommandDescription
}

func newPalette(commands []clicontract.CommandDescription) palette {
	in := textinput.New()
	in.Placeholder = "type to find a command"
	in.Focus()
	return palette{commands: commands, filter: in}
}

// matches are the commands the filter finds: names starting with it first,
// then names containing it, then descriptions mentioning it.
func (p palette) matches() []clicontract.CommandDescription {
	q := strings.ToLower(strings.TrimSpace(p.filter.Value()))
	var prefix, name, described []clicontract.CommandDescription
	for _, c := range p.commands {
		switch {
		case strings.HasPrefix(c.Name, q):
			prefix = append(prefix, c)
		case strings.Contains(c.Name, q):
			name = append(name, c)
		case strings.Contains(strings.ToLower(c.Description), q):
			described = append(described, c)
		}
	}
	return append(append(prefix, name...), described...)
}

// update handles a key; argv is set to the command line to run once the user
// confirms, closed is true when they back out of the palette.
func (p palette) update(msg tea.KeyMsg) (next palette, argv []string, closed bool) {
	switch msg.String() {
	case "esc", "ctrl+c":
		if p.chosen != nil {
			p.chosen = nil
			p.filter.SetValue("")
			p.filter.Placeholder = "type to find a command"
			return p, nil, false
		}
		return p, nil, true
	case "up":
		if p.cursor > 0 {
			p.cursor--
		}
		return p, nil, false
	case "down":
		if p.cursor < len(p.matches())-1 {
			p.cursor++
		}
		return p, nil, false
	case "enter":
		if p.chosen != nil {
			return p, append([]string{p.chosen.Name}, strings.Fields(p.filter.Value())...), false
		}
		found := p.matches()
		if len(found) == 0 {
			return p, nil, false
		}
		c := found[min(p.cursor, len(found)-1)]
		p.chosen = &c
		p.filter.SetValue("")
		p.filter.Placeholder = usage(c)
		return p, nil, false
	}
	p.filter, _ = p.filter.Update(msg)
	p.cursor = min(p.cursor, max(len(p.matches())-1, 0))
	return p, nil, false
}

// usage is a command's arguments and options as a placeholder.
func usage(c clicontract.CommandDescription) string {
	var parts []string
	for _, a := range c.Arguments {
		if a.Required {
			parts = append(parts, "<"+a.Name+">")
		} else {
			parts = append(parts, "["+a.Name+"]")
		}
	}
	for _, o := range c.Options {
		parts = append(parts, "[--"+o.Name+"]")
		if len(parts) > 6 {
			parts = append(parts, "…")
			break
		}
	}
	if len(parts) == 0 {
		return "enter to run"
	}
	return strings.Join(parts, " ")
}

func (p palette) view() string {
	var b strings.Builder
	if p.chosen != nil {
		b.WriteString(titleStyle.Render("smith "+p.chosen.Name) + " " + p.filter.View() + "\n")
		b.WriteString(hintStyle.Render("\nenter run · esc back") + "\n")
		return b.String()
	}
	b.WriteString(titleStyle.Render("› ") + p.filter.View() + "\n\n")
	for i, c := range p.matches() {
		if i >= 12 {
			break
		}
		b.WriteString(line(i == p.cursor, "", c.Name, c.Description))
	}
	b.WriteString(hintStyle.Render("\n↑/↓ choose · enter pick · esc close") + "\n")
	return b.String()
}
