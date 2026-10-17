package pubrelease

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/pubpackage"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// Request is the released app version to publish.
type Request struct {
	Project, Version string
}

// fetch downloads release assets from a tag's GitHub Release into the
// workspace; tests replace it.
var fetch = func(repo githubrelease.Repository, tag, root string, assets []string) error {
	client, err := githubrelease.FromEnvironment()
	if err != nil {
		return err
	}
	rel, found, err := client.ReleaseByTag(repo, tag)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("%s has no GitHub Release to take the archives from", tag)
	}
	byName := map[string]githubrelease.Asset{}
	for _, a := range rel.Assets {
		byName[a.Name] = a
	}
	for _, asset := range assets {
		a, ok := byName[path.Base(asset)]
		if !ok {
			return fmt.Errorf("the GitHub Release %s has no %s", tag, path.Base(asset))
		}
		if err := client.DownloadAsset(repo, a, filepath.Join(root, filepath.FromSlash(asset))); err != nil {
			return err
		}
	}
	return nil
}

// has reports whether pub.dev has a version; tests replace it.
var has = pubpackage.Has

// publishPackage builds and publishes the package; tests replace it.
var publishPackage = pubpackage.Publish

// Publish publishes an app version to pub.dev from its GitHub Release's
// archives, unless pub.dev already has it.
func Publish(s *session.Session, req Request) (pubpackage.Result, error) {
	var res pubpackage.Result
	p := s.Graph.Get(req.Project)
	if p == nil {
		return res, fmt.Errorf("no project %q", req.Project)
	}
	r := p.Release
	if r == nil || r.Pub == "" {
		return res, fmt.Errorf("%s is not published to pub.dev: set pub in [projects.%s.release]", p.Name, p.Name)
	}
	ok, err := has(r.Pub, req.Version)
	if err != nil {
		return res, fmt.Errorf("pub.dev: %w", err)
	}
	if ok {
		res.Existing = r.Pub + "@" + req.Version
		return res, nil
	}
	description, err := pubpackage.Description(r.Description)
	if err != nil {
		return res, fmt.Errorf("%s: %w", p.Name, err)
	}
	planned, err := release.PlannedAssets(s, p, req.Version)
	if err != nil {
		return res, err
	}
	repo, err := release.Repository(s)
	if err != nil {
		return res, err
	}
	archives := make([]string, 0, len(planned.AssetPlatforms))
	for asset := range planned.AssetPlatforms {
		archives = append(archives, asset)
	}
	sort.Strings(archives)
	if err := fetch(repo, release.TagPrefix(p)+req.Version, s.Root, archives); err != nil {
		return res, err
	}
	binary := planned.Binary
	if binary == "" {
		binary = p.Name
	}
	repository := "https://github.com/" + repo.String()
	homepage := r.Homepage
	if homepage == "" {
		homepage = repository
	}
	pr := pubpackage.Release{
		Name: r.Pub, Binary: binary, Version: req.Version, Description: description,
		Homepage: homepage, Repository: repository, License: "LICENSE",
		OutDir: "dist/release/" + p.Name + "/pub",
	}
	for _, asset := range archives {
		goos, goarch, _ := strings.Cut(planned.AssetPlatforms[asset], "/")
		pr.Archives = append(pr.Archives, pubpackage.Archive{Path: asset, OS: goos, Arch: goarch})
	}
	return publishPackage(s.Root, pr, false)
}
