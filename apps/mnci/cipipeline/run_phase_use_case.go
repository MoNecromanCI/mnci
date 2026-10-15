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

	"github.com/MoNecromanCI/mnci/apps/mnci/affected"
	"github.com/MoNecromanCI/mnci/apps/mnci/release"
	"github.com/MoNecromanCI/mnci/apps/mnci/session"
	"github.com/MoNecromanCI/mnci/apps/mnci/taskrun"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// VerifyTasks are what a pull request must pass.
var VerifyTasks = []string{"lint", "typecheck", "test", "build"}

// Phases lists `mnci ci` phases in pipeline order.
var Phases = []string{"setup", "audit", "verify", "release"}

// RunPhase runs one CI phase.
func RunPhase(ctx context.Context, s *session.Session, phase string, out io.Writer) error {
	begin := time.Now()
	group(out, "mnci ci "+phase)
	err := runPhase(ctx, s, phase, out)
	endGroup(out)
	status := "passed"
	if err != nil {
		status = "failed"
		annotate(out, phase, err)
	}
	fmt.Fprintf(out, "mnci ci %s %s in %s\n", phase, status, time.Since(begin).Round(time.Second))
	return err
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
		summary, err := taskrun.RunInWorkspace(ctx, s, taskrun.Request{Tasks: VerifyTasks, Affected: base != "", Base: base, Out: out})
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
			argv[i] = strings.ReplaceAll(a, "{workspaceRoot}", filepath.ToSlash(root))
		}
		code, err := taskrun.RunProcess(ctx, taskrun.Spec{Argv: argv, Dir: root, Env: os.Environ(), Out: out})
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
		_, _ = s.Git.Run("config", "user.email", "mnci-release@users.noreply.github.com")
	}
	if v, _ := s.Git.Run("config", "user.name"); v == "" {
		_, _ = s.Git.Run("config", "user.name", "mnci release")
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
		fmt.Fprintf(out, "::error title=mnci ci %s::%s\n", phase, msg)
	case os.Getenv("TF_BUILD") != "":
		fmt.Fprintf(out, "##vso[task.logissue type=error]mnci ci %s: %s\n", phase, msg)
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
