package tui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/MoNecromanCI/mnci/apps/mnci/pluginhost"
	"github.com/MoNecromanCI/mnci/apps/mnci/workspace"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

// Interactive reports whether a bare `mnci` should open the TUI.
func Interactive() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) && os.Getenv("CI") == ""
}

// Choose shows the menu and returns the mnci arguments the user built, or nil
// when they quit.
func Choose(version string, out io.Writer) ([]string, error) {
	header := "mnci " + version
	inside := false
	enabled := pluginhost.BuiltinNames()
	if root, err := workspace.FindRoot("."); err == nil {
		if cfg, err := workspace.Load(root); err == nil {
			inside = true
			header += " · " + cfg.Workspace.Name
			if len(cfg.Plugins.Enabled) > 0 {
				enabled = cfg.Plugins.Enabled
			}
		}
	}
	var actions []Action
	for _, a := range Actions(enabled, pluginhost.BuiltinNames(), enabled) {
		if (a.NeedsWorkspace && !inside) || (a.OutsideOnly && inside) {
			continue
		}
		actions = append(actions, a)
	}
	final, err := tea.NewProgram(NewModel(header, actions), tea.WithOutput(out)).Run()
	if err != nil {
		return nil, err
	}
	argv := final.(Model).Argv
	if len(argv) > 0 {
		fmt.Fprintf(out, "→ mnci %s\n\n", strings.Join(argv, " "))
	}
	return argv, nil
}
