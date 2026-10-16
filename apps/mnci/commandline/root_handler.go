package commandline

import (
	"errors"
	"fmt"
	"github.com/MoNecromanCI/mnci/apps/mnci/adoption"
	"github.com/MoNecromanCI/mnci/apps/mnci/pipelinetemplate"
	"os"

	"github.com/MoNecromanCI/mnci/apps/mnci/cipipeline"
	"github.com/MoNecromanCI/mnci/apps/mnci/clicontract"
	"github.com/MoNecromanCI/mnci/apps/mnci/configedit"
	"github.com/MoNecromanCI/mnci/apps/mnci/convert"
	"github.com/MoNecromanCI/mnci/apps/mnci/dependencies"
	"github.com/MoNecromanCI/mnci/apps/mnci/doctor"
	"github.com/MoNecromanCI/mnci/apps/mnci/explain"
	"github.com/MoNecromanCI/mnci/apps/mnci/graphview"
	"github.com/MoNecromanCI/mnci/apps/mnci/pluginhost"
	"github.com/MoNecromanCI/mnci/apps/mnci/release"
	"github.com/MoNecromanCI/mnci/apps/mnci/scaffold"
	"github.com/MoNecromanCI/mnci/apps/mnci/slicecheck"
	"github.com/MoNecromanCI/mnci/apps/mnci/stackview"
	"github.com/MoNecromanCI/mnci/apps/mnci/taskrun"
	"github.com/MoNecromanCI/mnci/apps/mnci/tui"
	"github.com/MoNecromanCI/mnci/apps/mnci/upgrade"
	"github.com/MoNecromanCI/mnci/apps/mnci/whymiss"
	"github.com/spf13/cobra"
)

// Root builds the mnci command.
func Root(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "mnci",
		Short: "A simple monorepo tool: add apps, libs and packages; build, test, release",
		Long: `mnci runs a monorepo in Node, Go and Python (and anything a plugin adds):
it scaffolds projects, runs their tasks in dependency order from a cache,
generates CI pipelines, and releases packages from conventional commits.`,
		Example: `  mnci                      open the interactive menu (in a terminal)
  mnci new acme
  mnci add app web --lang node
  mnci build`,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddGroup(&cobra.Group{ID: "start", Title: "Create:"}, &cobra.Group{ID: "work", Title: "Work:"}, &cobra.Group{ID: "ship", Title: "Ship:"}, &cobra.Group{ID: "inspect", Title: "Inspect:"})
	add := func(group string, cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.GroupID = group
			root.AddCommand(c)
		}
	}
	add("start", scaffold.NewCommand(version), scaffold.AddCommand(version), scaffold.RemoveCommand(version), convert.Command(version), configedit.Command(), upgrade.Command(version), adoption.Command(version))
	add("work", taskrun.Commands(version)...)
	add("work", dependencies.Commands(version)...)
	add("ship", release.Command(version), cipipeline.Command(version), pipelinetemplate.Command(version))
	add("inspect", graphview.Commands(version)...)
	add("inspect", clicontract.Commands(version)...)
	add("inspect", doctor.Command(version), explain.Command(version), whymiss.Command(version), pluginhost.Command(version), stackview.Command(version), slicecheck.Command())
	return root
}

// Execute runs mnci and returns the process exit code.
func Execute(version string, args []string) int {
	if len(args) == 0 && tui.Interactive() {
		chosen, err := tui.Choose(version, os.Stdout, clicontract.DescribeCommands(Root(version)))
		if err != nil {
			fmt.Fprintln(os.Stderr, "mnci: "+err.Error())
			return 1
		}
		if len(chosen) == 0 {
			return 0
		}
		args = chosen
	}
	root := Root(version)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		if !errors.Is(err, taskrun.ErrTasksFailed) && !errors.Is(err, doctor.ErrUnhealthy) && !errors.Is(err, adoption.ErrBlocked) {
			fmt.Fprintln(os.Stderr, "mnci: "+err.Error())
		}
		return 1
	}
	return 0
}
