package nativebuild

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/MoNecromanCI/mnci/apps/mnci/githubrelease"
	"github.com/MoNecromanCI/mnci/apps/mnci/projectgraph"
	"github.com/MoNecromanCI/mnci/apps/mnci/release"
	"github.com/MoNecromanCI/mnci/apps/mnci/session"
	"github.com/MoNecromanCI/mnci/apps/mnci/taskrun"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Tasks are what a native project runs on every runner.
var Tasks = []string{"lint", "test", "build", "package"}

// Run builds the native projects on this machine. On a Linux CI runner it
// first installs [ci] native_linux_packages. On a push to the release branch
// in GitHub Actions it then attaches this platform's build of every native app
// released at HEAD to the app's GitHub Release.
func Run(ctx context.Context, s *session.Session, out io.Writer) error {
	var native []*projectgraph.Project
	for _, p := range s.Graph.Projects {
		if release.Native(p) {
			native = append(native, p)
		}
	}
	if len(native) == 0 {
		fmt.Fprintln(out, "nothing to build natively: no project is tagged native")
		return nil
	}
	if packages := s.Config.CI.NativeLinuxPackages; len(packages) > 0 && runtime.GOOS == "linux" && inCI() {
		argv := append([]string{"sudo", "apt-get", "install", "-y", "--no-install-recommends"}, packages...)
		for _, cmd := range [][]string{{"sudo", "apt-get", "update"}, argv} {
			if code, err := taskrun.RunProcess(ctx, taskrun.Spec{Argv: cmd, Dir: s.Root, Env: os.Environ(), Out: out}); err != nil || code != 0 {
				return fmt.Errorf("%s failed (exit %d): %v", strings.Join(cmd, " "), code, err)
			}
		}
	}
	if err := os.MkdirAll(filepath.Join(s.Root, "dist", "drop"), 0o755); err != nil {
		return err
	}
	summary, err := taskrun.RunInWorkspace(ctx, s, taskrun.Request{Tasks: Tasks, Projects: []string{"tag:" + release.NativeTag}, Out: out})
	if err != nil {
		return err
	}
	fmt.Fprintln(out, summary.String())
	if !summary.OK() {
		return taskrun.ErrTasksFailed
	}
	if !releasePush(s) {
		return nil
	}
	return attach(s, native, out)
}

// attach adds this platform's assets to the GitHub Release of every native
// app tagged at HEAD.
func attach(s *session.Session, native []*projectgraph.Project, out io.Writer) error {
	head, err := s.Git.RevParse("HEAD")
	if err != nil {
		return err
	}
	for _, p := range native {
		if p.Release == nil {
			continue
		}
		tag, version := taggedAt(s, p, head)
		if tag == "" {
			fmt.Fprintf(out, "%s was not released by this commit; nothing to attach\n", p.Name)
			continue
		}
		spec := release.AssetSpec(p)
		spec.Platforms = []string{runtime.GOOS + "/" + runtime.GOARCH}
		var built protocol.PublishResult
		params := protocol.PublishParams{Project: p.Project, Version: version, Assets: spec}
		if ok, err := s.Plugins.Call(p.Plugin, protocol.MethodPublish, params, &built, 30*time.Minute); err != nil || !ok {
			return fmt.Errorf("building %s's %s asset: %v", p.Name, spec.Platforms[0], err)
		}
		client, err := githubrelease.FromEnvironment()
		if err != nil {
			return err
		}
		repo, err := release.Repository(s)
		if err != nil {
			return err
		}
		files := make([]string, len(built.Assets))
		for i, a := range built.Assets {
			files[i] = filepath.Join(s.Root, filepath.FromSlash(a))
		}
		res, err := githubrelease.Ensure(client, repo, githubrelease.Request{Tag: tag, Title: p.Name + " " + version, Assets: files})
		if err != nil {
			return fmt.Errorf("GitHub Release %s: %w", tag, err)
		}
		fmt.Fprintf(out, "attached %s to %s\n", strings.Join(res.Uploaded, ", "), tag)
	}
	return nil
}

// taggedAt is the project's release tag pointing at commit, and its version.
func taggedAt(s *session.Session, p *projectgraph.Project, commit string) (string, string) {
	prefix := release.TagPrefix(p)
	tags, _ := s.Git.Tags(prefix)
	for _, t := range tags {
		if c, err := s.Git.RevParse("refs/tags/" + t + "^{commit}"); err == nil && c == commit {
			return t, strings.TrimPrefix(t, prefix)
		}
	}
	return "", ""
}

func inCI() bool { return os.Getenv("GITHUB_ACTIONS") == "true" || os.Getenv("TF_BUILD") != "" }

// releasePush reports a GitHub Actions push to the release branch.
func releasePush(s *session.Session) bool {
	return os.Getenv("GITHUB_ACTIONS") == "true" && os.Getenv("GITHUB_EVENT_NAME") == "push" && os.Getenv("GITHUB_REF_NAME") == s.Config.Release.Branch
}
