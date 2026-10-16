package cipipeline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/affected"
	"github.com/monosmith-app/monosmith/apps/smith/nativebuild"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// VerifyTasks are what a pull request must pass.
var VerifyTasks = []string{"lint", "typecheck", "test", "build"}

// PackTask is the task that builds what a project deploys into dist/drop.
const PackTask = "package"

// DropDir is where packed deliverables go, workspace-relative.
const DropDir = "dist/drop"

// Phases lists `smith ci` phases in pipeline order.
var Phases = []string{"setup", "audit", "verify", "pack", "native", "release"}

// RunPhase runs one CI phase.
func RunPhase(ctx context.Context, s *session.Session, phase string, out io.Writer) error {
	begin := time.Now()
	group(out, "smith ci "+phase)
	err := runPhase(ctx, s, phase, out)
	endGroup(out)
	status := "passed"
	if err != nil {
		status = "failed"
		annotate(out, phase, err)
	}
	took := time.Since(begin).Round(100 * time.Millisecond)
	fmt.Fprintf(out, "smith ci %s %s in %s\n", phase, status, took)
	summarise(phase, err, took)
	return err
}

// summarise adds the phase's outcome to the GitHub Actions job summary.
func summarise(phase string, err error, took time.Duration) {
	file := os.Getenv("GITHUB_STEP_SUMMARY")
	if file == "" {
		return
	}
	line := fmt.Sprintf("- ✅ **%s** passed in %s\n", phase, took)
	if err != nil {
		line = fmt.Sprintf("- ❌ **%s** failed after %s\n", phase, took)
	}
	f, openErr := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if openErr != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = f.WriteString(line)
}

func runPhase(ctx context.Context, s *session.Session, phase string, out io.Writer) error {
	switch phase {
	case "setup", "audit":
		c, err := Collect(s)
		if err != nil {
			return err
		}
		steps := c.Setup
		if phase == "audit" {
			steps = c.Audit
		}
		if len(steps) == 0 {
			fmt.Fprintln(out, "nothing to "+phase)
		}
		return runSteps(ctx, s.Root, steps, out)
	case "verify":
		base := affected.BaseFromEnv()
		// Native projects are verified by the native job, on every OS.
		summary, err := taskrun.RunInWorkspace(ctx, s, taskrun.Request{Tasks: VerifyTasks, Projects: []string{"!tag:" + release.NativeTag}, Affected: base != "", Base: base, Out: out})
		if err != nil {
			return err
		}
		if len(summary.Results) > 0 {
			fmt.Fprintln(out, summary.String())
		}
		if !summary.OK() {
			return taskrun.ErrTasksFailed
		}
		return nil
	case "pack":
		// The directory always exists, so the artifact upload has a path.
		if err := os.MkdirAll(filepath.Join(s.Root, filepath.FromSlash(DropDir)), 0o755); err != nil {
			return err
		}
		packing := false
		for _, p := range s.Graph.Projects {
			if _, ok := p.Tasks[PackTask]; ok {
				packing = true
			}
		}
		if !packing {
			fmt.Fprintln(out, "nothing to pack: no project has a package task")
			return nil
		}
		// Native projects are packed by the native job, on each OS.
		summary, err := taskrun.RunInWorkspace(ctx, s, taskrun.Request{Tasks: []string{PackTask}, Projects: []string{"!tag:" + release.NativeTag}, Out: out})
		if err != nil {
			return err
		}
		fmt.Fprintln(out, summary.String())
		if !summary.OK() {
			return taskrun.ErrTasksFailed
		}
		fmt.Fprintf(out, "packed into %s\n", DropDir)
		return nil
	case "native":
		return nativebuild.Run(ctx, s, out)
	case "release":
		ensureIdentity(s)
		plan, err := release.MakePlan(s.Git, s.Graph)
		if err != nil {
			return err
		}
		report, err := release.Execute(ctx, s, plan, release.Options{Push: true, Out: out})
		fmt.Fprint(out, report.Markdown())
		return err
	}
	return fmt.Errorf("unknown phase %q (phases: %s)", phase, strings.Join(Phases, ", "))
}

func runSteps(ctx context.Context, root string, steps []protocol.Step, out io.Writer) error {
	var failures []string
	for _, step := range steps {
		fmt.Fprintf(out, "▶ %s\n", step.Name)
		argv := make([]string, len(step.Command))
		for i, a := range step.Command {
			argv[i] = strings.NewReplacer("{smith}", taskrun.Self(), "{workspaceRoot}", filepath.ToSlash(root)).Replace(a)
		}
		dir := root
		if step.Dir != "" {
			dir = filepath.Join(root, filepath.FromSlash(step.Dir))
		}
		code, err := taskrun.RunProcess(ctx, taskrun.Spec{Argv: argv, Dir: dir, Env: os.Environ(), Out: out})
		if err == nil && code == 0 {
			continue
		}
		msg := fmt.Sprintf("%s (exit %d)", step.Name, code)
		if err != nil {
			msg = step.Name + ": " + err.Error()
		}
		if step.Blocking {
			return errors.New(msg)
		}
		failures = append(failures, msg)
		warn(out, msg+" — not blocking")
	}
	if len(failures) > 0 {
		fmt.Fprintf(out, "%d non-blocking step(s) reported problems\n", len(failures))
	}
	return nil
}

// ensureIdentity gives CI a committer identity for annotated release tags.
func ensureIdentity(s *session.Session) {
	if v, _ := s.Git.Run("config", "user.email"); v == "" {
		_, _ = s.Git.Run("config", "user.email", "smith-release@users.noreply.github.com")
	}
	if v, _ := s.Git.Run("config", "user.name"); v == "" {
		_, _ = s.Git.Run("config", "user.name", "smith release")
	}
}

func group(out io.Writer, name string) {
	switch {
	case os.Getenv("GITHUB_ACTIONS") == "true":
		fmt.Fprintf(out, "::group::%s\n", name)
	case os.Getenv("TF_BUILD") != "":
		fmt.Fprintf(out, "##[group]%s\n", name)
	}
}

func endGroup(out io.Writer) {
	switch {
	case os.Getenv("GITHUB_ACTIONS") == "true":
		fmt.Fprintln(out, "::endgroup::")
	case os.Getenv("TF_BUILD") != "":
		fmt.Fprintln(out, "##[endgroup]")
	}
}

func annotate(out io.Writer, phase string, err error) {
	msg := strings.ReplaceAll(err.Error(), "\n", " ")
	switch {
	case os.Getenv("GITHUB_ACTIONS") == "true":
		fmt.Fprintf(out, "::error title=smith ci %s::%s\n", phase, msg)
	case os.Getenv("TF_BUILD") != "":
		fmt.Fprintf(out, "##vso[task.logissue type=error]smith ci %s: %s\n", phase, msg)
	}
}

func warn(out io.Writer, msg string) {
	switch {
	case os.Getenv("GITHUB_ACTIONS") == "true":
		fmt.Fprintf(out, "::warning::%s\n", msg)
	case os.Getenv("TF_BUILD") != "":
		fmt.Fprintf(out, "##vso[task.logissue type=warning]%s\n", msg)
	default:
		fmt.Fprintln(out, "warning: "+msg)
	}
}
