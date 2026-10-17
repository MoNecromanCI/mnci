package projectgraph

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestADependsOnOverrideKeepsATaskThePluginAddedForIt(t *testing.T) {
	p := &Project{Project: protocol.Project{Name: "core", Root: "packages/core", Tasks: map[string]protocol.Task{
		"build": {Command: []string{"tsc"}},
	}}}
	merge(p, protocol.Contribution{Replace: map[string]protocol.Task{
		"e2e":          {Command: []string{"jest", "-c", "e2e"}, DependsOn: []string{"e2e-browsers", "^build", "build"}},
		"e2e-browsers": {Command: []string{"npx", "playwright", "install"}},
	}}, "plugin playwright (convention playwright)")
	cfg := workspace.Config{Projects: map[string]workspace.ProjectOverride{"core": {Tasks: map[string]workspace.TaskOverride{
		"e2e": {DependsOn: []string{"^build"}},
	}}}}
	if err := ApplyOverrides(p, cfg); err != nil {
		t.Fatal(err)
	}
	got := p.Tasks["e2e"].DependsOn
	if strings.Join(got, ",") != "^build,e2e-browsers" {
		t.Errorf("depends_on = %v; the override's list plus e2e-browsers, which only the plugin's e2e needs", got)
	}
	if src := p.Sources["e2e.dependsOn"]; !strings.Contains(src, "[projects.core.tasks.e2e]") || !strings.Contains(src, "e2e-browsers from plugin playwright") {
		t.Errorf("source = %q", src)
	}
	if d := p.Dropped["e2e"]; len(d) != 1 || d[0] != "build (from plugin playwright (convention playwright))" {
		t.Errorf("dropped = %v; build is the user's to leave out, and doctor names it", d)
	}
	if len(cfg.Projects["core"].Tasks["e2e"].DependsOn) != 1 {
		t.Error("the override's own list was changed")
	}
}

func TestADependsOnOverrideOfALanguagePluginsTaskIsRecorded(t *testing.T) {
	p := &Project{Project: protocol.Project{Name: "web", Root: "apps/web", Tasks: map[string]protocol.Task{
		"test": {Command: []string{"jest"}, DependsOn: []string{"build"}},
		"lint": {Command: []string{"eslint", "."}},
	}}, Plugin: "node"}
	cfg := workspace.Config{Projects: map[string]workspace.ProjectOverride{"web": {Tasks: map[string]workspace.TaskOverride{
		"test": {DependsOn: []string{"lint"}},
		"lint": {DependsOn: []string{"^build"}},
	}}}}
	if err := ApplyOverrides(p, cfg); err != nil {
		t.Fatal(err)
	}
	if got := p.Tasks["test"].DependsOn; len(got) != 1 || got[0] != "lint" {
		t.Errorf("depends_on = %v: only a task a contribution added for this one is kept", got)
	}
	if d := p.Dropped["test"]; len(d) != 1 || d[0] != "build (from plugin node)" {
		t.Errorf("dropped = %v", d)
	}
	if _, ok := p.Dropped["lint"]; ok {
		t.Error("lint had no dependencies to drop")
	}
}

func TestApplyOverridesSplitRepo(t *testing.T) {
	project := func() *Project {
		return &Project{Project: protocol.Project{Name: "client", Root: "packages/client", Tasks: map[string]protocol.Task{}}, Sources: map[string]string{}}
	}
	p := project()
	cfg := workspace.Config{Projects: map[string]workspace.ProjectOverride{"client": {SplitRepo: "acme/client-php"}}}
	if err := ApplyOverrides(p, cfg); err != nil {
		t.Fatal(err)
	}
	if p.SplitRepo != "acme/client-php" || p.Release != nil {
		t.Errorf("split_repo is a project setting, not a release section (which builds assets): %+v", p)
	}

	cfg.Projects["client"] = workspace.ProjectOverride{SplitRepo: "https://github.com/acme/client-php"}
	err := ApplyOverrides(project(), cfg)
	if err == nil || !strings.Contains(err.Error(), "[projects.client]") || !strings.Contains(err.Error(), "owner/repo") {
		t.Errorf("a URL is refused, naming where it was set: %v", err)
	}
}
