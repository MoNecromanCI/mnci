package tui

import (
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"golang.org/x/term"
)

// Interactive reports whether a bare `smith` should open the TUI.
func Interactive() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) && os.Getenv("CI") == ""
}

// Choose shows the shell and returns the smith arguments the user built, or
// nil when they quit. commands feed the palette.
func Choose(version string, out io.Writer, commands []clicontract.CommandDescription) ([]string, error) {
	header := "smith " + version
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
	var sources *Sources
	if inside {
		s := WorkspaceSources(".", version)
		sources = &s
	}
	final, err := tea.NewProgram(NewShell(NewModel(header, actions), commands, sources), tea.WithOutput(out)).Run()
	if err != nil {
		return nil, err
	}
	argv := final.(Shell).Argv
	if len(argv) > 0 {
		fmt.Fprintf(out, "→ smith %s\n\n", strings.Join(argv, " "))
	}
	return argv, nil
}
