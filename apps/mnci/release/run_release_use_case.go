package release

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

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
		return nil
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

	if !opts.DryRun {
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

	var pr protocol.PublishResult
	ok, err = s.Plugins.Call(p.Plugin, protocol.MethodPublish, protocol.PublishParams{Project: p.Project, Version: item.Next, DryRun: opts.DryRun}, &pr, 15*time.Minute)
	if err != nil {
		return fmt.Errorf("publishing: %w", err)
	}
	if ok {
		res.Published, res.Registry, res.Publish = pr.Published, pr.Registry, pr.Skipped
	} else {
		res.Publish = "published by its tag"
	}
	if opts.DryRun {
		say("would release %s (%s)", item.Next, item.Reason)
		return nil
	}

	if err := s.Git.CreateTag(item.Tag, fmt.Sprintf("%s %s", item.Project, item.Next)); err != nil {
		return err
	}
	res.Status = "released"
	say("tagged %s", item.Tag)
	if opts.Push && s.Git.HasRemote(s.Config.Release.Remote) {
		if err := s.Git.PushTag(s.Config.Release.Remote, item.Tag); err != nil {
			return fmt.Errorf("pushing %s: %w (the tag exists locally; re-run to push it)", item.Tag, err)
		}
		res.Pushed = true
	}
	return nil
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
