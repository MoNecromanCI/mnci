//go:build e2e

package goldenpath

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestJvmWorkspace builds Maven modules with the built-in jvm plugin: an app
// depending on a package, every Maven task, the cache, dist/ for every
// build, and a release plan.
func TestJvmWorkspace(t *testing.T) {
	for _, tool := range []string{"java", "mvn"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip(tool + " is not installed")
		}
	}
	// The modules install into the user's Maven repository; leave it as found.
	if home, err := os.UserHomeDir(); err == nil {
		t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(home, ".m2", "repository", "dev", "mncie2ejvm")) })
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildMnci(t, base), dir: base, env: identity()}
	w.mnci(true, "new", "mnci-e2e-jvm", "--ci", "github")
	w.dir = filepath.Join(base, "mnci-e2e-jvm")
	w.mnci(true, "add", "app", "api", "--lang", "jvm")
	w.mnci(true, "add", "package", "core", "--lang", "jvm")

	// api uses core: a dependency on a workspace module is a graph edge.
	pom := "apps/api/pom.xml"
	w.write(pom, strings.Replace(w.read(pom), "  <dependencies>\n", "  <dependencies>\n    <dependency>\n      <groupId>dev.mncie2ejvm</groupId>\n      <artifactId>mnci-e2e-jvm-core</artifactId>\n      <version>0.0.0</version>\n    </dependency>\n", 1))
	main := "apps/api/src/main/java/dev/mncie2ejvm/api/Main.java"
	w.write(main, strings.Replace(w.read(main), `    System.out.println(GreetUseCase.greet("world"));`, `    String core = dev.mncie2ejvm.core.greeting.GreetUseCase.greet("core");
    System.out.println(GreetUseCase.greet("world") + " " + core);`, 1))
	if !hasEdge(w.mnci(true, "graph", "--format", "json"), "api", "core") {
		t.Fatal("a module dependency is a graph edge")
	}
	out := w.mnci(true, "run", "lint", "test", "build")
	for _, want := range []string{"api:build ✔", "core:test ✔", "api:test ✔", "core:lint ✔"} {
		if !strings.Contains(out, want) {
			t.Fatalf("%s did not pass:\n%s", want, out)
		}
	}
	if out := w.mnci(true, "run", "lint", "test"); !strings.Contains(out, "(4 from cache") {
		t.Fatalf("a re-run replays from the cache:\n%s", out)
	}
	if out := w.mnci(true, "run", "start", "-p", "api"); !strings.Contains(out, "Hello, world! Hello, core!") {
		t.Fatalf("start:\n%s", out)
	}
	// Builds go to dist/, never target/ in a source tree.
	for _, dir := range []string{"apps/api/target", "packages/core/target"} {
		if _, err := os.Stat(filepath.Join(w.dir, filepath.FromSlash(dir))); err == nil {
			t.Errorf("%s exists: Maven builds must go to dist/", dir)
		}
	}
	if _, err := os.Stat(filepath.Join(w.dir, "dist", "maven", "mnci-e2e-jvm-core")); err != nil {
		t.Error("the builds are under dist/maven")
	}
	if out := w.mnci(true, "doctor"); !strings.Contains(out, "✔ Maven is installed") || !strings.Contains(out, "✔ every build writes to dist/") {
		t.Fatalf("doctor:\n%s", out)
	}
	if ci := w.read(".github/workflows/ci.yml"); !strings.Contains(ci, "actions/setup-java") || !strings.Contains(ci, "MAVEN_REPOSITORY_URL") {
		t.Fatalf("ci.yml installs Java and passes the repository:\n%s", ci)
	}

	w.git("add", "-A")
	w.git("commit", "-qm", "feat: an api and its core")
	if plan := w.mnci(true, "release", "--dry-run"); !strings.Contains(plan, "core") || !strings.Contains(plan, "0.1.0") {
		t.Fatalf("release plan:\n%s", plan)
	}
}
