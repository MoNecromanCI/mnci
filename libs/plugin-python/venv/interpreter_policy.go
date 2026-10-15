package venv

import "runtime"

// Interpreter is the venv's python, as a command the host expands:
// {workspaceRoot} becomes the absolute workspace path at run time, so the cache
// key stays the same on every machine.
func Interpreter() string {
	if runtime.GOOS == "windows" {
		return "{workspaceRoot}/.venv/Scripts/python.exe"
	}
	return "{workspaceRoot}/.venv/bin/python"
}

// Launchers are the commands tried, in order, to find a system Python that can
// create the venv.
func Launchers() [][]string {
	if runtime.GOOS == "windows" {
		return [][]string{{"py", "-3"}, {"python"}, {"python3"}}
	}
	return [][]string{{"python3"}, {"python"}}
}
