//go:build e2e

package goldenpath

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestDockerWorkspace ships a Go app as a container image with the docker
// convention: the Dockerfile written when the app is added, the image
// task, and a release that pushes the versioned image to a registry.
func TestDockerWorkspace(t *testing.T) {
	out, err := exec.Command("docker", "version", "--format", "{{.Server.Os}}").Output()
	if err != nil || strings.TrimSpace(string(out)) != "linux" {
		t.Skip("no Docker daemon running Linux containers")
	}
	// A throwaway registry on a free port.
	registry := fmt.Sprintf("smith-e2e-registry-%d", freePort(t))
	if out, err := exec.Command("docker", "run", "-d", "--rm", "--name", registry, "-p", "127.0.0.1::5000", "registry:2").CombinedOutput(); err != nil {
		t.Fatalf("starting a registry: %v\n%s", err, out)
	}
	t.Cleanup(func() {
		_ = exec.Command("docker", "rm", "-f", registry).Run()
		_ = exec.Command("docker", "image", "rm", "-f", "api:dev", "api:latest").Run()
	})
	port, err := exec.Command("docker", "port", registry, "5000").Output()
	if err != nil {
		t.Fatal(err)
	}
	address := strings.TrimSpace(strings.Split(string(port), "\n")[0])

	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-docker", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-docker")
	w.write("smith.toml", w.read("smith.toml")+"\n[conventions.docker]\nregistry = \""+address+"/smith-e2e\"\n")
	w.smith(true, "add", "app", "api", "--lang", "go")
	if dockerfile := w.read("apps/api/Dockerfile"); !strings.Contains(dockerfile, "go build -trimpath -o /out/app ./apps/api") {
		t.Fatalf("the convention writes a new Go app's Dockerfile:\n%s", dockerfile)
	}
	if _, err := os.Stat(filepath.Join(w.dir, ".dockerignore")); err != nil {
		t.Fatal(".dockerignore keeps the context small")
	}
	if out := w.smith(true, "run", "image", "-p", "api"); !strings.Contains(out, "api:image ✔") {
		t.Fatalf("image:\n%s", out)
	}
	// A registry on this machine takes no login, so the pipeline asks for no
	// container credentials.
	if ci := w.read(".github/workflows/ci.yml"); strings.Contains(ci, "CONTAINER_") {
		t.Fatalf("ci.yml asks for credentials a local registry does not need:\n%s", ci)
	}

	// An app with a Dockerfile releases: its image is pushed with the version.
	w.git("add", "-A")
	w.git("commit", "-qm", "feat: the api")
	if out := w.smith(true, "release", "--dry-run"); !strings.Contains(out, "would push "+address+"/smith-e2e/api:0.1.0") {
		t.Fatalf("the dry run names the image it would push:\n%s", out)
	}
	if out := w.smith(true, "release", "--yes"); !strings.Contains(out, "docker published 0.1.0") {
		t.Fatalf("release:\n%s", out)
	}
	resp, err := http.Get("http://" + address + "/v2/smith-e2e/api/tags/list")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	tags, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(tags), `"0.1.0"`) || !strings.Contains(string(tags), `"latest"`) {
		t.Fatalf("the registry has the release's tags: %s", tags)
	}
}
