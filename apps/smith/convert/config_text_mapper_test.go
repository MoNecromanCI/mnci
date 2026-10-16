package convert

import (
	"reflect"
	"testing"

	"github.com/MoNecromanCI/mnci/apps/mnci/workspace"
)

func TestConfigTextRoundTrips(t *testing.T) {
	yes := true
	in := workspace.Config{
		Workspace: workspace.WorkspaceSection{Name: "acme", Scope: "@acme", ModulePrefix: "github.com/acme/repo"},
		CI:        workspace.CISection{Providers: []string{"github", "azure"}},
		Release:   workspace.ReleaseSection{GitHubReleases: true},
		Tasks:     map[string]workspace.TaskOverride{"test": {DependsOn: []string{"^build"}}},
		Projects: map[string]workspace.ProjectOverride{
			"tool": {
				Release: &workspace.ProjectRelease{AssetName: "go-app-{product}-{os}-{arch}.{ext}", Product: "tool", OSAlias: map[string]string{"darwin": "macos"}},
				Tasks: map[string]workspace.TaskOverride{
					"slice-check": {Command: []string{"node", "tools/check.cjs", `C:\x "y"`}, Cwd: ".", Inputs: []string{"**/*.go"}, Cache: &yes, Env: map[string]string{"A_B": "1"}},
				},
			},
			"web.app": {TagPrefix: "@acme/web@"},
		},
	}
	text := ConfigText(in)
	out, err := workspace.Parse([]byte(text))
	if err != nil {
		t.Fatalf("%v\n%s", err, text)
	}
	if out.Workspace != in.Workspace || !reflect.DeepEqual(out.CI.Providers, in.CI.Providers) || !out.Release.GitHubReleases {
		t.Fatalf("sections %+v\n%s", out, text)
	}
	tool := out.Projects["tool"]
	if tool.Release.Product != "tool" || tool.Release.OSAlias["darwin"] != "macos" || len(tool.Release.Platforms) != 6 {
		t.Fatalf("release %+v", tool.Release)
	}
	check := tool.Tasks["slice-check"]
	if !reflect.DeepEqual(check.Command, in.Projects["tool"].Tasks["slice-check"].Command) || check.Cwd != "." || !*check.Cache || check.Env["A_B"] != "1" {
		t.Fatalf("task %+v", check)
	}
	if out.Projects["web.app"].TagPrefix != "@acme/web@" || out.Tasks["test"].DependsOn[0] != "^build" {
		t.Fatalf("%+v", out)
	}
}
