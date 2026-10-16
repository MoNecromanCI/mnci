package projectgraph

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/MoNecromanCI/mnci/apps/mnci/workspace"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Project is a plugin's project plus what mnci adds to it.
type Project struct {
	protocol.Project
	// Plugin is the name of the plugin that discovered the project.
	Plugin string `json:"plugin"`
	// Deps are names of projects this one imports, sorted.
	Deps []string `json:"deps"`
	// Sources maps "task.field" to where its value came from
	// ("plugin node", "mnci.toml [tasks.test]", ...).
	Sources map[string]string `json:"-"`
	// Release is set for an app that releases (mnci.toml
	// [projects.<name>.release]).
	Release *workspace.ProjectRelease `json:"release,omitempty"`
	// Conventions are the conventions that apply to the project, sorted.
	Conventions []string `json:"conventions,omitempty"`
	// contributed maps "task.field" to the plugin or convention that set it.
	contributed map[string]string
	// stackPlugins are the plugins of the project's stack, other than its
	// language plugin.
	stackPlugins []string
}

// Releases reports whether the project takes part in releases: every
// package, and an app that opted in.
func (p *Project) Releases() bool {
	return p.Kind == protocol.ProjectPackage || (p.Kind == protocol.ProjectApp && p.Release != nil)
}

// ServedBy reports whether a plugin owns the project, is part of its stack,
// or applies to it as a convention: the projects a plugin's ci and doctor
// answers cover.
func (p *Project) ServedBy(plugin string) bool {
	return p.Plugin == plugin || slices.Contains(p.stackPlugins, plugin) || slices.Contains(p.Conventions, plugin)
}

// Graph is every project and the edges between them.
type Graph struct {
	Projects []*Project
	byName   map[string]*Project
}

// New indexes projects by name. Names must be unique.
func New(projects []*Project) (*Graph, error) {
	g := &Graph{byName: map[string]*Project{}}
	sort.Slice(projects, func(i, j int) bool { return projects[i].Name < projects[j].Name })
	for _, p := range projects {
		if other, dup := g.byName[p.Name]; dup {
			return nil, fmt.Errorf("two projects are named %q (%s and %s); project names must be unique — rename one directory", p.Name, other.Root, p.Root)
		}
		g.byName[p.Name] = p
		g.Projects = append(g.Projects, p)
	}
	return g, nil
}

// Get returns a project by name, or nil.
func (g *Graph) Get(name string) *Project { return g.byName[name] }

// Names lists project names, sorted.
func (g *Graph) Names() []string {
	out := make([]string, len(g.Projects))
	for i, p := range g.Projects {
		out[i] = p.Name
	}
	return out
}

// Dependents returns the names of projects that import name directly.
func (g *Graph) Dependents(name string) []string {
	var out []string
	for _, p := range g.Projects {
		for _, d := range p.Deps {
			if d == name {
				out = append(out, p.Name)
			}
		}
	}
	return out
}

// TransitiveDeps returns every project name reachable from name, excluding it.
func (g *Graph) TransitiveDeps(name string) []string {
	return g.walk(name, func(p *Project) []string { return p.Deps })
}

// TransitiveDependents returns every project that reaches name, excluding it.
func (g *Graph) TransitiveDependents(name string) []string {
	return g.walk(name, func(p *Project) []string { return g.Dependents(p.Name) })
}

func (g *Graph) walk(start string, next func(*Project) []string) []string {
	seen := map[string]bool{start: true}
	queue := []string{start}
	var out []string
	for len(queue) > 0 {
		p := g.byName[queue[0]]
		queue = queue[1:]
		if p == nil {
			continue
		}
		for _, n := range next(p) {
			if !seen[n] {
				seen[n] = true
				out = append(out, n)
				queue = append(queue, n)
			}
		}
	}
	sort.Strings(out)
	return out
}

// Order returns project names so that every project comes after the projects
// it imports. It fails on a cycle and names it.
func (g *Graph) Order() ([]string, error) {
	const (
		unvisited = iota
		visiting
		done
	)
	state := map[string]int{}
	var order []string
	var stack []string
	var visit func(string) error
	visit = func(name string) error {
		switch state[name] {
		case done:
			return nil
		case visiting:
			i := 0
			for stack[i] != name {
				i++
			}
			return fmt.Errorf("projects import each other in a cycle: %s", strings.Join(append(stack[i:], name), " → "))
		}
		state[name] = visiting
		stack = append(stack, name)
		for _, d := range g.byName[name].Deps {
			if err := visit(d); err != nil {
				return err
			}
		}
		stack = stack[:len(stack)-1]
		state[name] = done
		order = append(order, name)
		return nil
	}
	for _, p := range g.Projects {
		if err := visit(p.Name); err != nil {
			return nil, err
		}
	}
	return order, nil
}

// Owner returns the project whose root contains a workspace-relative path, the
// deepest one when roots nest, or nil.
func (g *Graph) Owner(path string) *Project {
	var best *Project
	for _, p := range g.Projects {
		if path == p.Root || strings.HasPrefix(path, p.Root+"/") {
			if best == nil || len(p.Root) > len(best.Root) {
				best = p
			}
		}
	}
	return best
}
