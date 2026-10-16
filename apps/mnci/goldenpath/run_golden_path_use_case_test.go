//go:build e2e

package goldenpath

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

type workspace struct {
	t   *testing.T
	bin string
	dir string
	env []string
}

// mnci runs the binary and returns its output; wantOK says whether it must succeed.
func (w *workspace) mnci(wantOK bool, args ...string) string {
	w.t.Helper()
	cmd := exec.Command(w.bin, args...)
	cmd.Dir = w.dir
	cmd.Env = append(os.Environ(), w.env...)
	began := time.Now()
	out, err := cmd.CombinedOutput()
	w.t.Logf("mnci %s (%s)", strings.Join(args, " "), time.Since(began).Round(time.Millisecond))
	if (err == nil) != wantOK {
		w.t.Fatalf("mnci %s: err=%v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func (w *workspace) write(rel, body string) {
	w.t.Helper()
	p := filepath.Join(w.dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		w.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		w.t.Fatal(err)
	}
}

func (w *workspace) git(args ...string) string {
	w.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = w.dir
	cmd.Env = append(os.Environ(), w.env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		w.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func buildMnci(t *testing.T, dir string) string {
	t.Helper()
	bin := filepath.Join(dir, "mnci")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	wd, _ := os.Getwd()
	run(t, filepath.Dir(wd), "go", "build", "-ldflags", "-X main.version=0.0.0-e2e", "-o", bin, ".")
	return bin
}

func TestGoldenPath(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildMnci(t, base), dir: base}
	for _, kv := range [][2]string{{"user.email", "e2e@example.com"}, {"user.name", "e2e"}} {
		w.env = append(w.env, "GIT_AUTHOR_"+strings.ToUpper(strings.TrimPrefix(kv[0], "user."))+"="+kv[1], "GIT_COMMITTER_"+strings.ToUpper(strings.TrimPrefix(kv[0], "user."))+"="+kv[1])
	}

	// 1. A new workspace is committed and CI-ready.
	w.mnci(true, "new", "acme", "--ci", "github,azure")
	w.dir = filepath.Join(base, "acme")
	for _, f := range []string{"mnci.toml", ".github/workflows/ci.yml", "azure-pipelines.yml"} {
		if _, err := os.Stat(filepath.Join(w.dir, filepath.FromSlash(f))); err != nil {
			t.Fatalf("new did not write %s", f)
		}
	}

	// 2. An app, a lib and a package in each language. Each project picks its
	// stack: the language default (node-jest), a published one (node-vitest),
	// or one mnci.toml declares (go-sliced).
	toml, _ := os.ReadFile(filepath.Join(w.dir, "mnci.toml"))
	w.write("mnci.toml", string(toml)+"\n[stacks.go-sliced]\nlanguage = \"go\"\nstructure = [\"vertical-slices\"]\n")
	for _, add := range [][]string{
		{"app", "web", "--lang", "node"}, {"lib", "ui", "--lang", "node"}, {"package", "sdk", "--stack", "node-vitest"}, {"app", "hello", "--lang", "node", "--stack", "none"},
		{"app", "api", "--lang", "go"}, {"lib", "core", "--lang", "go"}, {"package", "gokit", "--stack", "go-sliced"},
		{"app", "worker", "--lang", "python"}, {"lib", "pyutil", "--lang", "python"}, {"package", "pykit", "--lang", "python"},
	} {
		w.mnci(true, append([]string{"add"}, add...)...)
	}
	for dir, want := range map[string][]string{
		"apps/web/package.json":     {`"stack": "node-jest"`, `"test": "jest"`, `"lint": "eslint ."`},
		"packages/sdk/package.json": {`"stack": "node-vitest"`, `"test": "vitest run"`, `"lint": "eslint ."`},
		"packages/gokit/go.mod":     {"// mnci:stack go-sliced"},
		"apps/hello/package.json":   {`"stack": "none"`, `"start": "node src/main.js"`},
	} {
		manifest, _ := os.ReadFile(filepath.Join(w.dir, filepath.FromSlash(dir)))
		for _, s := range want {
			if !strings.Contains(string(manifest), s) {
				t.Fatalf("%s lacks %s:\n%s", dir, s, manifest)
			}
		}
	}
	if stacks := w.mnci(true, "stacks"); !strings.Contains(stacks, "node-vitest  (plugin vitest)") || !strings.Contains(stacks, "projects   sdk") || !strings.Contains(stacks, "go-sliced  (mnci.toml)") {
		t.Fatalf("stacks:\n%s", stacks)
	}
	projects := w.mnci(true, "projects", "--json")
	var list []struct{ Name, Kind, Language string }
	if err := json.Unmarshal([]byte(projects[strings.Index(projects, "["):]), &list); err != nil || len(list) != 10 {
		t.Fatalf("projects: %v\n%s", err, projects)
	}

	// 3. Imports between projects work immediately.
	w.write("apps/web/src/main.ts", "import { greet as ui } from '@acme/ui'\nimport { greet } from './greeting'\n\nprocess.stdout.write(greet('world').message + ' ' + ui('ui').message + '\\n')\n")
	w.write("apps/api/main.go", "package main\n\nimport (\n\t\"fmt\"\n\n\t\"example.com/acme/apps/api/greeting\"\n\tcore \"example.com/acme/libs/core/greeting\"\n)\n\nfunc main() {\n\tfmt.Println(greeting.Greet(\"world\").Message, core.Greet(\"core\").Message)\n}\n")
	w.write("apps/worker/src/acme_worker/__main__.py", "from acme_pyutil.greeting import greet as util\n\nfrom .greeting import greet\n\nprint(greet(\"world\").message, util(\"util\").message)\n")
	graph := w.mnci(true, "graph", "--format", "json")
	for _, edge := range [][2]string{{"web", "ui"}, {"api", "core"}, {"worker", "pyutil"}} {
		if !strings.Contains(graph, fmt.Sprintf(`"name": %q`, edge[0])) || !hasEdge(graph, edge[0], edge[1]) {
			t.Fatalf("graph lacks %s → %s:\n%s", edge[0], edge[1], graph)
		}
	}

	// 4. Everything lints, type-checks, tests and builds; a re-run is all cache.
	out := w.mnci(true, "run", "lint", "typecheck", "test", "build")
	if !strings.Contains(out, "tasks succeeded") {
		t.Fatal(out)
	}
	// The vertical-slices convention checks Go and Python layouts before lint.
	for _, want := range []string{"api:slice-check", "pykit:slice-check"} {
		if !strings.Contains(out, want) {
			t.Fatalf("%s did not run:\n%s", want, out)
		}
	}
	w.write("libs/core/helpers/format.go", "package helpers\n")
	if out := w.mnci(false, "run", "slice-check", "-p", "core"); !strings.Contains(out, `"helpers" is a junk drawer`) {
		t.Fatalf("a junk-drawer folder passes the slice check:\n%s", out)
	}
	if err := os.RemoveAll(filepath.Join(w.dir, "libs", "core", "helpers")); err != nil {
		t.Fatal(err)
	}
	out = w.mnci(true, "run", "lint", "typecheck", "test", "build")
	if strings.Contains(out, "▶") {
		t.Fatalf("second run executed tasks:\n%s", out)
	}
	// why-miss names what would make a task run again.
	if out := w.mnci(true, "why-miss", "api:build"); !strings.Contains(out, "would come from the cache") {
		t.Fatalf("why-miss on a cached task:\n%s", out)
	}
	w.write("libs/core/greeting/extra.go", "package greeting\n")
	if out := w.mnci(true, "why-miss", "api:build"); !strings.Contains(out, "upstream   core changed") {
		t.Fatalf("why-miss must name the changed upstream project:\n%s", out)
	}
	if err := os.Remove(filepath.Join(w.dir, "libs", "core", "greeting", "extra.go")); err != nil {
		t.Fatal(err)
	}
	out = w.mnci(true, "run", "start", "-p", "web,api,worker,hello", "--parallel", "1")
	for _, want := range []string{"Hello, ui!", "Hello, core!", "Hello, util!", "hello:start │ Hello, world!"} {
		if !strings.Contains(out, want) {
			t.Fatalf("apps must run with their internal deps (%s):\n%s", want, out)
		}
	}
	w.mnci(true, "doctor")
	w.mnci(true, "ci", "generate", "--check")

	// The phases the generated pipeline runs, exactly as CI runs them.
	for _, phase := range []string{"setup", "audit", "verify"} {
		if out := w.mnci(true, "ci", phase); !strings.Contains(out, "mnci ci "+phase+" passed") {
			t.Fatalf("ci %s:\n%s", phase, out)
		}
	}

	// 5. Affected follows dependents and stays narrow.
	w.git("add", "-A")
	w.git("commit", "-qm", "feat: add every kind of project")
	w.git("branch", "base")
	w.write("libs/ui/src/greeting/extra.ts", "export const extra = 1\n")
	affected := w.mnci(true, "affected", "--base", "base")
	if !strings.Contains(affected, "ui") || !strings.Contains(affected, "web") || strings.Contains(affected, "api") {
		t.Fatalf("affected:\n%s", affected)
	}
	_ = os.Remove(filepath.Join(w.dir, "libs", "ui", "src", "greeting", "extra.ts"))

	// 6. Release: preview, a release interrupted halfway, and a resume.
	reg := startRegistries(t, base)
	remote := filepath.Join(base, "remote.git")
	run(t, base, "git", "init", "-q", "--bare", remote)
	w.git("remote", "add", "origin", remote)
	npmrc, _ := os.ReadFile(filepath.Join(w.dir, ".npmrc"))
	host := strings.TrimPrefix(reg.npmURL, "http:")
	w.write(".npmrc", string(npmrc)+"registry="+reg.npmURL+"\n"+host+":_authToken=${NODE_AUTH_TOKEN}\n")
	w.git("add", "-A")
	w.git("commit", "-qm", "ci: publish to the test registries")
	w.git("push", "-q", "origin", "main")

	preview := w.mnci(true, "release", "--dry-run")
	for _, p := range []string{"gokit", "pykit", "sdk"} {
		if !strings.Contains(preview, p) {
			t.Fatalf("preview lacks %s:\n%s", p, preview)
		}
	}
	w.env = append(w.env, "NODE_AUTH_TOKEN="+reg.npmToken, "TWINE_USERNAME=ci", "TWINE_PASSWORD=ci", "TWINE_REPOSITORY_URL=http://localhost:1/")
	failed := w.mnci(false, "release", "--yes")
	if !strings.Contains(failed, "Stopped") || !strings.Contains(w.git("ls-remote", "--tags", "origin"), "packages/gokit/v0.1.0") {
		t.Fatalf("the release must stop at pykit, keeping gokit:\n%s", failed)
	}
	if status := w.git("status", "--porcelain"); strings.TrimSpace(status) != "" {
		t.Fatalf("a failed release must restore manifests:\n%s", status)
	}
	w.env[len(w.env)-1] = "TWINE_REPOSITORY_URL=" + reg.pypiURL
	resumed := w.mnci(true, "release", "--yes")
	if !strings.Contains(resumed, "Unchanged: gokit") {
		t.Fatalf("the resume must keep gokit:\n%s", resumed)
	}
	tags := w.git("ls-remote", "--tags", "origin")
	for _, tag := range []string{"packages/gokit/v0.1.0", "pykit@0.1.0", "sdk@0.1.0"} {
		if !strings.Contains(tags, tag) {
			t.Fatalf("remote lacks %s:\n%s", tag, tags)
		}
	}
	if body := get(t, reg.npmURL+"@acme%2fsdk"); !strings.Contains(body, `"0.1.0"`) {
		t.Fatalf("npm registry lacks @acme/sdk@0.1.0:\n%s", body)
	}
	if body := get(t, reg.pypiURL+"simple/acme-pykit/"); !strings.Contains(body, "acme_pykit-0.1.0") {
		t.Fatalf("index lacks acme-pykit 0.1.0:\n%s", body)
	}
	if again := w.mnci(true, "release", "--yes"); !strings.Contains(again, "Nothing to release") {
		t.Fatalf("a third run releases nothing:\n%s", again)
	}

	// A published version whose tag is gone (here and on the remote) would
	// restart the package from its manifest: release refuses, even a dry run,
	// and doctor reports it, until the tag is back.
	commit := strings.TrimSpace(w.git("rev-parse", "sdk@0.1.0^{commit}"))
	w.git("tag", "-d", "sdk@0.1.0")
	w.git("push", "-q", "origin", ":refs/tags/sdk@0.1.0")
	if out := w.mnci(false, "release", "--dry-run"); !strings.Contains(out, "sdk is published at 0.1.0 but has no release tag") {
		t.Fatalf("a published, untagged package must block the release:\n%s", out)
	}
	if out := w.mnci(false, "doctor"); !strings.Contains(out, "release tags are not behind the registry") {
		t.Fatalf("doctor must report it:\n%s", out)
	}
	w.git("tag", "sdk@0.1.0", commit)
	w.git("push", "-q", "origin", "sdk@0.1.0")
	if out := w.mnci(true, "doctor"); strings.Contains(out, "published at") {
		t.Fatalf("doctor still reports the restored tag:\n%s", out)
	}

	// 7. A fix releases a patch.
	w.write("packages/sdk/src/fix.ts", "export const fixed = true\n")
	w.git("add", "-A")
	w.git("commit", "-qm", "fix(sdk): a patch")
	if plan := w.mnci(true, "release", "--dry-run"); !strings.Contains(plan, "0.1.0 → 0.1.1") {
		t.Fatalf("a fix must plan a patch:\n%s", plan)
	}
}

func hasEdge(graphJSON, from, to string) bool {
	var g struct {
		Projects []struct {
			Name string   `json:"name"`
			Deps []string `json:"deps"`
		} `json:"projects"`
	}
	if json.Unmarshal([]byte(graphJSON), &g) != nil {
		return false
	}
	for _, p := range g.Projects {
		if p.Name == from {
			for _, d := range p.Deps {
				if d == to {
					return true
				}
			}
		}
	}
	return false
}

func get(t *testing.T, url string) string {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var b strings.Builder
	buf := make([]byte, 32<<10)
	for {
		n, err := resp.Body.Read(buf)
		b.Write(buf[:n])
		if err != nil {
			break
		}
	}
	return b.String()
}
