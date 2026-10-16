//go:build !windows

package taskrun

import (
	"os/exec"
	"syscall"
	"time"
)

// prepare puts the child in its own process group so the whole tree can be
// signalled.
func prepare(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// track has nothing to do on Unix: the process group is the tree.
func track(*exec.Cmd) (release func(), err error) { return func() {}, nil }

// killTree sends SIGTERM to the group, then SIGKILL after a grace period.
func killTree(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	pgid := -cmd.Process.Pid
	_ = syscall.Kill(pgid, syscall.SIGTERM)
	time.AfterFunc(3*time.Second, func() { _ = syscall.Kill(pgid, syscall.SIGKILL) })
}
