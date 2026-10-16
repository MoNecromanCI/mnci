package tui

import (
	"strings"

	"github.com/MoNecromanCI/mnci/apps/mnci/clicontract"
	tea "github.com/charmbracelet/bubbletea"
)

type tab int

const (
	tabMenu tab = iota
	tabCommands
	tabDoctor
	tabRelease
	tabSettings
)

var tabNames = map[tab]string{tabMenu: "Menu", tabCommands: "Commands", tabDoctor: "Doctor", tabRelease: "Release", tabSettings: "Settings"}

// Shell is the TUI: the guided menu, a palette of every command, and, inside
// a workspace, the doctor, release and settings views. Nothing happens only here:
// every action ends as an mnci command line, printed before it runs.
type Shell struct {
	header   string
	tabs     []tab
	active   int
	menu     Model
	palette  palette
	doctor   doctorView
	release  releaseView
	settings settingsView
	sources  Sources
	// Argv is the chosen command, empty when the user quit.
	Argv []string
}

// NewShell builds the shell; without sources (outside a workspace) it has no
// doctor or release view.
func NewShell(menu Model, commands []clicontract.CommandDescription, sources *Sources) Shell {
	sh := Shell{header: menu.header, menu: menu, palette: newPalette(commands), tabs: []tab{tabMenu, tabCommands}}
	if sources != nil {
		sh.sources = *sources
		sh.tabs = append(sh.tabs, tabDoctor, tabRelease, tabSettings)
	}
	return sh
}

// Init implements tea.Model.
func (sh Shell) Init() tea.Cmd { return nil }

func (sh Shell) current() tab { return sh.tabs[sh.active] }

func (sh Shell) run(argv ...string) (tea.Model, tea.Cmd) {
	sh.Argv = argv
	return sh, tea.Quit
}

// Update implements tea.Model.
func (sh Shell) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case doctorLoaded:
		sh.doctor = doctorView{report: &msg.report}
		return sh, nil
	case doctorFailed:
		sh.doctor = doctorView{err: msg.err}
		return sh, nil
	case settingsLoaded:
		sh.settings = settingsView{loaded: true, settings: msg.settings, err: msg.err}
		return sh, nil
	case releaseLoaded:
		sh.release = releaseView{loaded: true, plan: msg.plan, blockers: msg.blockers, warnings: msg.warnings, err: msg.err}
		return sh, nil
	case tea.KeyMsg:
		return sh.key(msg)
	}
	return sh, nil
}

func (sh Shell) key(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return sh, tea.Quit
	case "tab", "shift+tab":
		step := 1
		if msg.String() == "shift+tab" {
			step = len(sh.tabs) - 1
		}
		sh.active = (sh.active + step) % len(sh.tabs)
		return sh, sh.enter()
	}
	switch sh.current() {
	case tabMenu:
		next, cmd := sh.menu.Update(msg)
		sh.menu = next.(Model)
		if sh.menu.done {
			return sh.run(sh.menu.Argv...)
		}
		return sh, cmd
	case tabCommands:
		next, argv, closed := sh.palette.update(msg)
		sh.palette = next
		if argv != nil {
			return sh.run(argv...)
		}
		if closed {
			sh.active = 0
		}
		return sh, nil
	case tabDoctor:
		switch msg.String() {
		case "q", "esc":
			return sh, tea.Quit
		case "r":
			sh.doctor = doctorView{}
			return sh, sh.enter()
		case "f":
			return sh.run("doctor", "--fix")
		}
		sh.doctor = sh.doctor.update(msg)
	case tabSettings:
		if !sh.settings.editing && (msg.String() == "q" || msg.String() == "esc") {
			return sh, tea.Quit
		}
		next, argv := sh.settings.update(msg)
		sh.settings = next
		if argv != nil {
			return sh.run(argv...)
		}
	case tabRelease:
		switch msg.String() {
		case "q", "esc":
			return sh, tea.Quit
		case "p":
			return sh.run("release", "--dry-run")
		case "R":
			if sh.release.loaded && sh.release.err == nil && len(sh.release.blockers) == 0 && len(sh.release.plan.Items) > 0 {
				return sh.run("release")
			}
		}
	}
	return sh, nil
}

type doctorFailed struct{ err error }

// enter loads the active view the first time it is shown.
func (sh Shell) enter() tea.Cmd {
	switch sh.current() {
	case tabDoctor:
		if sh.doctor.report == nil && sh.doctor.err == nil {
			load := sh.sources.Doctor
			return func() tea.Msg {
				report, err := load()
				if err != nil {
					return doctorFailed{err}
				}
				return doctorLoaded{report}
			}
		}
	case tabRelease:
		if !sh.release.loaded {
			load := sh.sources.Release
			return func() tea.Msg { return load() }
		}
	case tabSettings:
		if !sh.settings.loaded {
			load := sh.sources.Settings
			return func() tea.Msg { return load() }
		}
	}
	return nil
}

// View implements tea.Model.
func (sh Shell) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(sh.header) + "  ")
	for i, t := range sh.tabs {
		if i == sh.active {
			b.WriteString(cursorStyle.Render("["+tabNames[t]+"]") + " ")
		} else {
			b.WriteString(hintStyle.Render(" "+tabNames[t]+" ") + " ")
		}
	}
	b.WriteString(hintStyle.Render("· tab switch · ctrl+c quit") + "\n\n")
	switch sh.current() {
	case tabMenu:
		b.WriteString(sh.menu.body())
	case tabCommands:
		b.WriteString(sh.palette.view())
	case tabDoctor:
		b.WriteString(sh.doctor.view())
	case tabRelease:
		b.WriteString(sh.release.view())
	case tabSettings:
		b.WriteString(sh.settings.view())
	}
	return b.String()
}
