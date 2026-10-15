package release

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/MoNecromanCI/mnci/apps/mnci/githubrelease"
	"github.com/MoNecromanCI/mnci/apps/mnci/projectgraph"
	"github.com/MoNecromanCI/mnci/apps/mnci/session"
	"github.com/MoNecromanCI/mnci/apps/mnci/taskrun"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Options configure a release run.
type Options struct {
	DryRun bool
	// Push sends tags to the release remote (when it exists).
	Push bool
	Out  io.Writer
}

// Execute releases a plan. Each package is stamped, built, published, restored
// and tagged in dependency order; it stops at the first failure, keeping what
// already finished.
func Execute(ctx context.Context, s *session.Session, plan Plan, opts Options) (Report, error) {
	report := Report{DryRun: opts.DryRun, Started: time.Now().UTC(), Unchanged: plan.Unchanged, Results: []ReportResult{}}
	defer func() { report.Finished = time.Now().UTC(); writeReport(s.Root, report) }()

	if !opts.DryRun {
		if err := Preflight(s, plan); err != nil {
			report.Error = err.Error()
			return report, err
		}
		pushPending(s, opts)
	}
	versions := map[string]string{}
	for _, item := range plan.Items {
		for _, e := range s.Graph.Get(item.Project).Exports {
			versions[e] = item.Next
		}
	}
	for _, item := range plan.Items {
		res := ReportResult{Item: item, Status: "planned"}
		if err := releaseOne(ctx, s, item, versions, opts, &res); err != nil {
			res.Status, res.Error = "failed", err.Error()
			report.Results = append(report.Results, res)
			report.Error = item.Project + ": " + err.Error()
			return report, errors.New(report.Error)
		}
		report.Results = append(report.Results, res)
	}
	return report, nil
}

func releaseOne(ctx context.Context, s *session.Session, item Item, versions map[string]string, opts Options, res *ReportResult) error {
	p := s.Graph.Get(item.Project)
	say := func(format string, args ...any) { fmt.Fprintf(opts.Out, item.Project+": "+format+"\n", args...) }
	if commit := s.Git.TagCommit(item.Tag); commit != "" {
		res.Status = "already-released"
		say("%s already exists", item.Tag)
		if opts.DryRun || !s.Config.Release.GitHubReleases {
			return nil
		}
		if err := pushTag(s, item, opts, res); err != nil {
			return err
		}
		return publishGitHubRelease(s, p, item, res, say, nil)
	}

	deps := map[string]string{}
	for _, d := range s.Graph.TransitiveDeps(p.Name) {
		for _, e := range s.Graph.Get(d).Exports {
			if v, ok := versions[e]; ok {
				deps[e] = v
			}
		}
	}
	var vr protocol.VersionResult
	ok, err := s.Plugins.Call(p.Plugin, protocol.MethodVersion, protocol.VersionParams{Project: p.Project, Version: item.Next, Dependencies: deps, DryRun: opts.DryRun}, &vr, time.Minute)
	if err != nil {
		return fmt.Errorf("writing version: %w", err)
	}
	if ok {
		res.Files = vr.Files
	}
	if !opts.DryRun && len(vr.Files) > 0 {
		defer func() {
			if err := s.Git.RestoreFiles(vr.Files...); err != nil {
				say("could not restore %v: %v", vr.Files, err)
			}
		}()
	}

	// An app's assets are built by publishing; its own build is not needed.
	if !opts.DryRun && p.Release == nil {
		if _, has := p.Tasks["build"]; has {
			say("building %s", item.Next)
			summary, err := taskrun.RunInWorkspace(ctx, s, taskrun.Request{Tasks: []string{"build"}, Projects: []string{p.Name}, Out: opts.Out, NoCache: len(vr.Files) > 0})
			if err != nil {
				return err
			}
			if !summary.OK() {
				return errors.New("build failed")
			}
		}
	}

	if p.Release != nil && !opts.DryRun {
		say("building assets for %s", item.Next)
	}
	pr, err := publish(s, p, item, opts.DryRun)
	if err != nil {
		return err
	}
	res.Published, res.Registry, res.Publish, res.Assets = pr.Published, pr.Registry, pr.Skipped, pr.Assets
	if opts.DryRun {
		say("would release %s (%s)", item.Next, item.Reason)
		if len(pr.Assets) > 0 {
			say("would attach %s", strings.Join(pr.Assets, ", "))
		}
		return nil
	}

	if err := s.Git.CreateTag(item.Tag, fmt.Sprintf("%s %s", item.Project, item.Next)); err != nil {
		return err
	}
	res.Status = "released"
	say("tagged %s", item.Tag)
	if err := pushTag(s, item, opts, res); err != nil {
		return err
	}
	if s.Config.Release.GitHubReleases {
		return publishGitHubRelease(s, p, item, res, say, pr.Assets)
	}
	return nil
}

// publish asks the project's plugin to publish it; an app with a release
// configuration gets its assets built instead.
func publish(s *session.Session, p *projectgraph.Project, item Item, dryRun bool) (protocol.PublishResult, error) {
	params := protocol.PublishParams{Project: p.Project, Version: item.Next, DryRun: dryRun, Assets: AssetSpec(p)}
	var pr protocol.PublishResult
	ok, err := s.Plugins.Call(p.Plugin, protocol.MethodPublish, params, &pr, 30*time.Minute)
	if err != nil {
		return pr, fmt.Errorf("publishing: %w", err)
	}
	if !ok {
		if params.Assets != nil {
			return pr, fmt.Errorf("the %s plugin cannot build release assets", p.Plugin)
		}
		pr.Skipped = "published by its tag"
	}
	return pr, nil
}

// AssetSpec is what a plugin needs to build an app's release assets, or nil
// for a project without them.
func AssetSpec(p *projectgraph.Project) *protocol.AssetSpec {
	r := p.Release
	if r == nil {
		return nil
	}
	return &protocol.AssetSpec{
		Name:            r.AssetName,
		Product:         r.Product,
		Platforms:       r.Platforms,
		OSAlias:         r.OSAlias,
		ArchAlias:       r.ArchAlias,
		VersionVariable: r.VersionVariable,
		OutDir:          "dist/release/" + p.Name,
	}
}

func pushTag(s *session.Session, item Item, opts Options, res *ReportResult) error {
	if !opts.Push || !s.Git.HasRemote(s.Config.Release.Remote) {
		return nil
	}
	if err := s.Git.PushTag(s.Config.Release.Remote, item.Tag); err != nil {
		return fmt.Errorf("pushing %s: %w (the tag exists locally; re-run to push it)", item.Tag, err)
	}
	res.Pushed = true
	return nil
}

// publishGitHubRelease creates the tag's GitHub Release and attaches the
// assets. When assets is nil (a resumed release), the plugin lists the
// expected files and they are rebuilt only if the release lacks any.
func publishGitHubRelease(s *session.Session, p *projectgraph.Project, item Item, res *ReportResult, say func(string, ...any), assets []string) error {
	if !res.Pushed {
		say("not creating a GitHub Release: the tag was not pushed")
		return nil
	}
	client, err := githubrelease.FromEnvironment()
	if err != nil {
		return err
	}
	repo, err := Repository(s)
	if err != nil {
		return err
	}
	if assets == nil && p.Release != nil {
		existing, found, err := client.ReleaseByTag(repo, item.Tag)
		if err != nil {
			return err
		}
		planned, err := publish(s, p, item, true)
		if err != nil {
			return err
		}
		if !found || missing(existing, planned.Assets) {
			say("building assets for %s", item.Next)
			built, err := publish(s, p, item, false)
			if err != nil {
				return err
			}
			assets = built.Assets
		}
	}
	files := make([]string, len(assets))
	for i, a := range assets {
		files[i] = filepath.Join(s.Root, filepath.FromSlash(a))
	}
	out, err := githubrelease.Ensure(client, repo, githubrelease.Request{Tag: item.Tag, Title: item.Project + " " + item.Next, Notes: githubrelease.Notes(item.Reason, item.Commits), Assets: files})
	if err != nil {
		return fmt.Errorf("GitHub Release %s: %w (re-run the release to finish it)", item.Tag, err)
	}
	res.GitHubRelease = out.URL
	if out.Created {
		say("created GitHub Release %s", out.URL)
	}
	if len(out.Uploaded) > 0 {
		say("attached %s", strings.Join(out.Uploaded, ", "))
	}
	return nil
}

func missing(r githubrelease.Release, assets []string) bool {
	have := map[string]bool{}
	for _, a := range r.Assets {
		have[a.Name] = true
	}
	for _, a := range assets {
		if !have[path.Base(a)] {
			return true
		}
	}
	return false
}

// Repository is the GitHub repository releases go to: GITHUB_REPOSITORY on
// Actions, otherwise the release remote's URL.
func Repository(s *session.Session) (githubrelease.Repository, error) {
	if env := os.Getenv("GITHUB_REPOSITORY"); env != "" {
		return githubrelease.ParseRepository(env)
	}
	remote, err := s.Git.RemoteURL(s.Config.Release.Remote)
	if err != nil {
		return githubrelease.Repository{}, fmt.Errorf("no GitHub repository: set GITHUB_REPOSITORY or add the %s remote", s.Config.Release.Remote)
	}
	return githubrelease.ParseRepository(remote)
}

// pushPending pushes release tags at HEAD left local by an interrupted run.
func pushPending(s *session.Session, opts Options) {
	if !opts.Push || !s.Git.HasRemote(s.Config.Release.Remote) {
		return
	}
	head, err := s.Git.RevParse("HEAD")
	if err != nil {
		return
	}
	out, err := s.Git.Run("tag", "--points-at", head)
	if err != nil || out == "" {
		return
	}
	for _, tag := range strings.Fields(out) {
		_ = s.Git.PushTag(s.Config.Release.Remote, tag)
	}
}

func writeReport(root string, r Report) {
	dir := filepath.Join(root, ".mnci")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	raw, err := json.MarshalIndent(r, "", "  ")
	if err == nil {
		_ = os.WriteFile(filepath.Join(dir, "release-report.json"), append(raw, '\n'), 0o644)
	}
	if summary := os.Getenv("GITHUB_STEP_SUMMARY"); summary != "" {
		if f, err := os.OpenFile(summary, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644); err == nil {
			_, _ = f.WriteString(r.Markdown() + "\n")
			_ = f.Close()
		}
	}
}
