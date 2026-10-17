package projectgraph

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

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
