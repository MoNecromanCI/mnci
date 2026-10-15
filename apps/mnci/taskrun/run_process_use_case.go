package taskrun

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Spec is one command to run.
type Spec struct {
	Argv []string
	Dir  string
	Env  []string
	Out  io.Writer
}

// RunProcess runs a command to completion and returns its exit code. When ctx
// is cancelled the whole process tree is stopped.
func RunProcess(ctx context.Context, spec Spec) (int, error) {
	if len(spec.Argv) == 0 {
		return -1, errors.New("task has no command")
	}
	program := spec.Argv[0]
	if !strings.ContainsAny(program, `/\`) {
		resolved, err := exec.LookPath(program)
		if err != nil {
			return -1, fmt.Errorf("%s is not installed or not on PATH", program)
		}
		program = resolved
	} else {
		program = filepath.FromSlash(program)
		if !filepath.IsAbs(program) {
			program = filepath.Join(spec.Dir, program)
		}
	}
	cmd := exec.Command(program, spec.Argv[1:]...)
	cmd.Dir = spec.Dir
	cmd.Env = spec.Env
	cmd.Stdout = spec.Out
	cmd.Stderr = spec.Out
	cmd.Stdin = nil
	prepare(cmd)
	if err := cmd.Start(); err != nil {
		return -1, err
	}
	release, _ := track(cmd)
	defer release()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return exitCode(err)
	case <-ctx.Done():
		killTree(cmd)
		<-done
		return -1, ctx.Err()
	}
}

func exitCode(err error) (int, error) {
	if err == nil {
		return 0, nil
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode(), nil
	}
	return -1, err
}

// Environment builds a child environment: the caller's, plus extra entries,
// replacing keys case-insensitively on Windows where Path and PATH are one.
func Environment(extra map[string]string) []string {
	env := os.Environ()
	for k, v := range extra {
		replaced := false
		for i, kv := range env {
			name, _, _ := strings.Cut(kv, "=")
			if name == k || (isWindows && strings.EqualFold(name, k)) {
				env[i] = name + "=" + v
				replaced = true
			}
		}
		if !replaced {
			env = append(env, k+"="+v)
		}
	}
	return env
}
