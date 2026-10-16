package releaselineage

import (
	"strings"
	"testing"
)

func TestAssess(t *testing.T) {
	tags := []string{
		"sdk@0.1.0", "sdk@0.2.0", "sdk@0.3.0-rc.1",
		"@acme/web@1.4.0", "web@1.2.0",
		"ui@2.0.0", "@acme/ui@1.0.0",
		"api@1.0.0",
		"packages/gokit/v0.3.0",
	}
	subjects := []Subject{
		// Tags only under the scoped package name: stranded.
		{Project: "web", Prefix: "web@", Alternatives: []string{"@acme/web@"}, Tags: tags, Published: "1.4.0"},
		// Its own tags are ahead of the old name: fine.
		{Project: "ui", Prefix: "ui@", Alternatives: []string{"@acme/ui@"}, Tags: tags, Published: "2.0.0"},
		// Tagged 0.2.0, registry has 0.4.0: behind.
		{Project: "sdk", Prefix: "sdk@", Tags: tags, Published: "0.4.0"},
		// Published but never tagged.
		{Project: "kit", Prefix: "kit@", Tags: tags, Published: "3.1.0"},
		// Not published yet: a first release, fine.
		{Project: "new", Prefix: "new@", Tags: tags},
		// Equal to the registry: fine.
		{Project: "api", Prefix: "api@", Tags: tags, Published: "1.0.0"},
		// Go: no registry answer, only its own tags.
		{Project: "gokit", Prefix: "packages/gokit/v", Tags: tags},
		// The registry could not be asked: a warning.
		{Project: "py", Prefix: "py@", Tags: tags, Unreachable: true, Registry: "https://pypi.org/simple/", Detail: "connection refused"},
	}
	got := map[string]Finding{}
	for _, f := range Assess(subjects) {
		got[f.Project] = f
	}
	if len(got) != 4 {
		t.Fatalf("findings for %v", got)
	}
	for project, want := range map[string]struct {
		blocking bool
		text     string
	}{
		"web": {true, `newest release tag is @acme/web@1.4.0, under another name than its prefix web@, so the next release would start from web@1.2.0. Set [projects.web] tag_prefix = "@acme/web@"`},
		"sdk": {true, "tagged sdk@0.2.0 but the registry already has 0.4.0"},
		"kit": {true, "published at 3.1.0 but has no release tag"},
		"py":  {false, "could not ask https://pypi.org/simple/ about py (connection refused)"},
	} {
		f := got[project]
		if f.Blocking != want.blocking || !strings.Contains(f.Message, want.text) {
			t.Errorf("%s: %+v\nwant %q", project, f, want.text)
		}
	}
	if !strings.Contains(got["web"].Message, "git tag web@1.4.0 @acme/web@1.4.0^{commit}") {
		t.Errorf("web remedy: %s", got["web"].Message)
	}
}
