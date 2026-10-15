package toolrun

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Log is where tool output goes; tests may replace it.
var Log io.Writer = os.Stderr

// Run executes argv in dir, streaming its output to Log. extraEnv entries are
// KEY=VALUE pairs added to the plugin's environment.
func Run(dir string, argv []string, extraEnv ...string) error {
	cmd, err := command(dir, argv, extraEnv)
	if err != nil {
		return err
	}
	cmd.Stdout = Log
	cmd.Stderr = Log
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %w", strings.Join(argv, " "), err)
	}
	return nil
}

// Output executes argv in dir and returns its stdout; stderr goes to Log.
func Output(dir string, argv []string, extraEnv ...string) (string, error) {
	cmd, err := command(dir, argv, extraEnv)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = Log
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("%s failed: %w", strings.Join(argv, " "), err)
	}
	return out.String(), nil
}

// Available reports whether a program can be found on PATH (honouring
// PATHEXT on Windows, so "npm" finds npm.cmd).
func Available(program string) bool {
	_, err := exec.LookPath(program)
	return err == nil
}

func command(dir string, argv []string, extraEnv []string) (*exec.Cmd, error) {
	if len(argv) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return nil, fmt.Errorf("%s is not installed or not on PATH", argv[0])
	}
	cmd := exec.Command(path, argv[1:]...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), extraEnv...)
	return cmd, nil
}
