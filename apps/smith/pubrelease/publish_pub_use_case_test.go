package pubrelease

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/pubpackage"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func workspaceWith(t *testing.T, release *workspace.ProjectRelease) *session.Session {
	t.Helper()
	g, err := projectgraph.New([]*projectgraph.Project{{
		Project: protocol.Project{Name: "smith", Root: "apps/smith", Kind: protocol.ProjectApp},
		Release: release,
	}})
	if err != nil {
		t.Fatal(err)
	}
	return &session.Session{Root: t.TempDir(), Graph: g}
}

func stub(t *testing.T, pubDevHas bool) (fetched, published *bool) {
	t.Helper()
	keepHas, keepFetch, keepPublish := has, fetch, publishPackage
	t.Cleanup(func() { has, fetch, publishPackage = keepHas, keepFetch, keepPublish })
	fetched, published = new(bool), new(bool)
	has = func(string, string) (bool, error) { return pubDevHas, nil }
	fetch = func(githubrelease.Repository, string, string, []string) error { *fetched = true; return nil }
	publishPackage = func(string, pubpackage.Release, bool) (pubpackage.Result, error) {
		*published = true
		return pubpackage.Result{}, nil
	}
	return fetched, published
}

func TestPublishSkipsAVersionPubDevHas(t *testing.T) {
	fetched, published := stub(t, true)
	s := workspaceWith(t, &workspace.ProjectRelease{Pub: "monosmith", Description: strings.Repeat("a long enough description ", 4)})
	res, err := Publish(s, Request{Project: "smith", Version: "1.4.0"})
	if err != nil || res.Existing != "monosmith@1.4.0" || *fetched || *published {
		t.Errorf("res %+v, err %v, fetched %v, published %v", res, err, *fetched, *published)
	}
}

func TestPublishRefusesWhatCannotGoToPubDev(t *testing.T) {
	fetched, published := stub(t, false)
	cases := map[string]struct {
		release *workspace.ProjectRelease
		project string
		want    string
	}{
		"unknown project":   {&workspace.ProjectRelease{Pub: "monosmith"}, "nope", `no project "nope"`},
		"no release":        {nil, "smith", "set pub in [projects.smith.release]"},
		"no pub":            {&workspace.ProjectRelease{NPM: []string{"x"}}, "smith", "set pub in [projects.smith.release]"},
		"short description": {&workspace.ProjectRelease{Pub: "monosmith", Description: "Runs tasks"}, "smith", "60 to 180 characters"},
	}
	for name, c := range cases {
		_, err := Publish(workspaceWith(t, c.release), Request{Project: c.project, Version: "1.4.0"})
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err = %v, want %q", name, err, c.want)
		}
	}
	if *fetched || *published {
		t.Error("fetched or published a refused package")
	}
}
