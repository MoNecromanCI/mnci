package scaffold

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/MoNecromanCI/mnci/apps/mnci/cipipeline"
	"github.com/MoNecromanCI/mnci/apps/mnci/projectgraph"
	"github.com/MoNecromanCI/mnci/apps/mnci/session"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/projectkind"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// AddOptions configure mnci add.
type AddOptions struct {
	Kind string
	Name string
	// Lang selects the plugin; empty picks the workspace's only language.
	Lang string
	Out  io.Writer
}

// AddProject generates a project through its language plugin, lets each
// configured convention that applies to it add its part (lint, tests), and
// regenerates the pipelines so CI knows about the new toolchain at once.
func AddProject(s *session.Session, opts AddOptions) (*protocol.GenerateResult, error) {
	kind := protocol.ProjectKind(opts.Kind)
	switch kind {
	case protocol.ProjectApp, protocol.ProjectLib, protocol.ProjectPackage:
	default:
		return nil, fmt.Errorf("kind %q is not app, lib or package", opts.Kind)
	}
	if err := ValidateName(opts.Kind, opts.Name); err != nil {
		return nil, err
	}
	if s.Graph.Get(opts.Name) != nil {
		return nil, fmt.Errorf("a project named %q already exists (%s)", opts.Name, s.Graph.Get(opts.Name).Root)
	}
	lang, err := chooseLanguage(s, opts.Lang)
	if err != nil {
		return nil, err
	}
	plugin, err := s.Plugins.ForLanguage(lang)
	if err != nil {
		return nil, err
	}
	if !plugin.Info.Capabilities.Generate {
		return nil, fmt.Errorf("plugin %s cannot generate projects", plugin.Info.Plugin.Name)
	}
	dir := projectkind.Directory(kind) + "/" + opts.Name
	if _, err := os.Stat(filepath.Join(s.Root, filepath.FromSlash(dir))); err == nil {
		return nil, fmt.Errorf("%s already exists", dir)
	}
	fmt.Fprintf(opts.Out, "Adding %s %s (%s) in %s\n", opts.Kind, opts.Name, plugin.Info.Plugin.Name, dir)
	var res protocol.GenerateResult
	params := protocol.GenerateParams{Kind: kind, Name: opts.Name, Directory: dir, Workspace: protocol.WorkspaceInfo{
		Name: s.Config.Workspace.Name, Scope: s.Config.Workspace.Scope, ModulePrefix: s.Config.Workspace.ModulePrefix,
	}}
	if err := plugin.Call(protocol.MethodGenerate, params, &res, 20*time.Minute); err != nil {
		return nil, err
	}
	if err := s.Refresh(); err != nil {
		return &res, err
	}
	project := s.Graph.Get(opts.Name)
	if project == nil {
		return &res, errors.New("the plugin generated files, but no project named " + opts.Name + " was discovered")
	}
	if err := generateConventions(s, project, params, &res, opts.Out); err != nil {
		return &res, err
	}
	changed, _, err := cipipeline.Regenerate(s)
	if err != nil {
		return &res, err
	}
	res.Files = append(res.Files, changed...)
	sort.Strings(res.Files)
	return &res, nil
}

func chooseLanguage(s *session.Session, lang string) (string, error) {
	if lang != "" {
		return lang, nil
	}
	used := map[string]bool{}
	for _, p := range s.Graph.Projects {
		used[p.Plugin] = true
	}
	if len(used) == 1 {
		for l := range used {
			return l, nil
		}
	}
	return "", fmt.Errorf("choose a language with --lang (%s)", strings.Join(s.Plugins.Languages(), ", "))
}

// generateConventions asks every configured convention that applies to the
// new project and can generate to add its files, then rediscovers.
func generateConventions(s *session.Session, project *projectgraph.Project, params protocol.GenerateParams, res *protocol.GenerateResult, out io.Writer) error {
	ran := false
	for _, c := range s.Plugins.Clients {
		name := c.Info.Plugin.Name
		conv, configured := s.Config.Conventions[name]
		if !configured || c.Info.Kind != protocol.KindConvention || !c.Supports(protocol.MethodGenerate) {
			continue
		}
		targets, err := projectgraph.ConventionTargets(name, []*projectgraph.Project{project}, c.Info.Languages, conv)
		if err != nil {
			return err
		}
		if len(targets) == 0 {
			continue
		}
		fmt.Fprintf(out, "Applying convention %s\n", name)
		var added protocol.GenerateResult
		if err := c.Call(protocol.MethodGenerate, params, &added, 20*time.Minute); err != nil {
			return fmt.Errorf("convention %s: %w", name, err)
		}
		res.Files = append(res.Files, added.Files...)
		res.Notes = append(res.Notes, added.Notes...)
		ran = true
	}
	if !ran {
		return nil
	}
	return s.Refresh()
}
