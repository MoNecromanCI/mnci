package projectgraph

import (
	"fmt"
	"sort"

	"github.com/MoNecromanCI/mnci/apps/mnci/workspace"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// TaskFields are the fields of a task whose source is tracked.
var TaskFields = []string{"command", "inputs", "outputs", "dependsOn", "env", "cache"}

// ApplyOverrides layers mnci.toml on top of plugin-inferred tasks: plugin
// defaults, then [tasks.<task>], then [projects.<name>.tasks.<task>]. A
// project override may also add a task the plugin does not know about, if it
// gives a command. Every field's source is recorded in Sources.
func ApplyOverrides(p *Project, cfg workspace.Config) error {
	p.Sources = map[string]string{}
	if p.Tasks == nil {
		p.Tasks = map[string]protocol.Task{}
	}
	for name := range p.Tasks {
		for _, f := range TaskFields {
			p.Sources[name+"."+f] = "plugin " + p.Plugin
		}
	}
	names := make([]string, 0, len(p.Tasks))
	for name := range p.Tasks {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if o, ok := cfg.Tasks[name]; ok {
			p.Tasks[name] = apply(p.Tasks[name], o, p.Sources, name, fmt.Sprintf("mnci.toml [tasks.%s]", name))
		}
	}
	if po, ok := cfg.Projects[p.Name]; ok {
		taskNames := make([]string, 0, len(po.Tasks))
		for name := range po.Tasks {
			taskNames = append(taskNames, name)
		}
		sort.Strings(taskNames)
		for _, name := range taskNames {
			o := po.Tasks[name]
			t, exists := p.Tasks[name]
			if !exists && len(o.Command) == 0 {
				return fmt.Errorf("mnci.toml [projects.%s.tasks.%s]: %s has no %q task, so the override must give a command", p.Name, name, p.Name, name)
			}
			p.Tasks[name] = apply(t, o, p.Sources, name, fmt.Sprintf("mnci.toml [projects.%s.tasks.%s]", p.Name, name))
		}
	}
	return nil
}

func apply(t protocol.Task, o workspace.TaskOverride, sources map[string]string, name, source string) protocol.Task {
	set := func(field string) { sources[name+"."+field] = source }
	if o.Command != nil {
		t.Command = o.Command
		set("command")
	}
	if o.Inputs != nil {
		t.Inputs = o.Inputs
		set("inputs")
	}
	if o.Outputs != nil {
		t.Outputs = o.Outputs
		set("outputs")
	}
	if o.DependsOn != nil {
		t.DependsOn = o.DependsOn
		set("dependsOn")
	}
	if o.Env != nil {
		t.Env = o.Env
		set("env")
	}
	if o.Cache != nil {
		t.Cache = *o.Cache
		set("cache")
	}
	return t
}
