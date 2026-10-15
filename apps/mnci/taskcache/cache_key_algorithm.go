package taskcache

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"sort"
)

// KeySchema versions the key encoding. Changing what goes into a key, or how,
// requires a new schema so old entries can never be mistaken for new ones.
const KeySchema = "mnci-task-key-v1"

// FileDigest is one input file: workspace-relative path and content digest.
type FileDigest struct {
	Path, Digest string
}

// KeyInputs is everything that can change a task's result.
type KeyInputs struct {
	Task string
	// Command is the task's argv before placeholders are expanded, so the key
	// does not depend on where the workspace is checked out.
	Command []string
	Env     map[string]string
	Files   []FileDigest
	// Upstream are digests of the sources of every project this one depends
	// on, transitively, by project name.
	Upstream map[string]string
	// DepKeys are the keys of tasks this one depends on.
	DepKeys  []string
	Platform string
	// Tool is the version of mnci and of the plugin that defined the task.
	Tool string
}

// Key computes the cache key: a SHA-256 over a length-prefixed encoding of
// each field, with every collection sorted.
func Key(in KeyInputs) string {
	h := sha256.New()
	write := func(s string) {
		var n [8]byte
		binary.BigEndian.PutUint64(n[:], uint64(len(s)))
		h.Write(n[:])
		h.Write([]byte(s))
	}
	section := func(name string, count int) {
		write(name)
		write(string(rune(count)))
	}
	write(KeySchema)
	write(in.Task)
	write(in.Platform)
	write(in.Tool)
	section("command", len(in.Command))
	for _, a := range in.Command {
		write(a)
	}
	section("env", len(in.Env))
	for _, k := range sortedKeys(in.Env) {
		write(k)
		write(in.Env[k])
	}
	files := append([]FileDigest(nil), in.Files...)
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	section("files", len(files))
	for _, f := range files {
		write(f.Path)
		write(f.Digest)
	}
	section("upstream", len(in.Upstream))
	for _, k := range sortedKeys(in.Upstream) {
		write(k)
		write(in.Upstream[k])
	}
	deps := append([]string(nil), in.DepKeys...)
	sort.Strings(deps)
	section("deps", len(deps))
	for _, d := range deps {
		write(d)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Combine digests a list of file digests into one, for a project's sources.
func Combine(files []FileDigest) string {
	return Key(KeyInputs{Task: "sources", Files: files})
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
