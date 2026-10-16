package releasing

import (
	"archive/zip"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestPublishModuleIsTheTag(t *testing.T) {
	res, err := Publish(t.TempDir(), protocol.PublishParams{Project: protocol.Project{Name: "lib"}, Version: "1.0.0"})
	if err != nil || res.Skipped != "published by its tag" {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestBuildAssets(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go unavailable")
	}
	root := t.TempDir()
	app := filepath.Join(root, "apps", "hello")
	_ = os.MkdirAll(app, 0o755)
	_ = os.WriteFile(filepath.Join(app, "go.mod"), []byte("module example.com/hello\n\ngo 1.26.0\n"), 0o644)
	_ = os.WriteFile(filepath.Join(app, "main.go"), []byte("package main\n\nvar version = \"dev\"\n\nfunc main() { println(version) }\n"), 0o644)
	spec := &protocol.AssetSpec{Name: "{product}_{version}_{os}_{arch}.{ext}", Platforms: []string{"linux/amd64", "windows/arm64"}, OutDir: "dist/release/hello"}
	params := protocol.PublishParams{Project: protocol.Project{Name: "hello", Root: "apps/hello"}, Version: "1.2.3", Assets: spec}

	params.DryRun = true
	planned, err := Publish(root, params)
	if err != nil || len(planned.Assets) != 3 || planned.Published {
		t.Fatalf("dry run %+v %v", planned, err)
	}
	if _, err := os.Stat(filepath.Join(root, "dist")); !os.IsNotExist(err) {
		t.Fatal("a dry run builds nothing")
	}

	params.DryRun = false
	res, err := Publish(root, params)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"dist/release/hello/hello_1.2.3_linux_amd64.zip", "dist/release/hello/hello_1.2.3_windows_arm64.zip", "dist/release/hello/checksums.txt"}
	if strings.Join(res.Assets, ",") != strings.Join(want, ",") || strings.Join(planned.Assets, ",") != strings.Join(want, ",") {
		t.Fatalf("assets %v, planned %v", res.Assets, planned.Assets)
	}
	z, err := zip.OpenReader(filepath.Join(root, "dist/release/hello/hello_1.2.3_windows_arm64.zip"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = z.Close() }()
	if len(z.File) != 1 || z.File[0].Name != "hello.exe" {
		t.Fatalf("zip holds %v", z.File)
	}
	sums, _ := os.ReadFile(filepath.Join(root, "dist/release/hello/checksums.txt"))
	if lines := strings.Split(strings.TrimSpace(string(sums)), "\n"); len(lines) != 2 || !strings.HasSuffix(lines[0], "  hello_1.2.3_linux_amd64.zip") {
		t.Fatalf("checksums %q", sums)
	}
	entries, _ := os.ReadDir(filepath.Join(root, "dist/release/hello"))
	if len(entries) != 3 {
		t.Fatalf("only the assets remain, got %v", entries)
	}
}

// TestBuildAssetsIsReproducible builds the same app in two different
// directories: -trimpath, a fixed zip header and no build-time values make
// the archives byte-identical, so anyone can check a release against its
// source.
func TestBuildAssetsIsReproducible(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go unavailable")
	}
	build := func() string {
		root := t.TempDir()
		app := filepath.Join(root, "apps", "hello")
		_ = os.MkdirAll(app, 0o755)
		_ = os.WriteFile(filepath.Join(app, "go.mod"), []byte("module example.com/hello\n\ngo 1.26.0\n"), 0o644)
		_ = os.WriteFile(filepath.Join(app, "main.go"), []byte("package main\n\nvar version = \"dev\"\n\nfunc main() { println(version) }\n"), 0o644)
		spec := &protocol.AssetSpec{Name: "{product}_{version}_{os}_{arch}.{ext}", Platforms: []string{"linux/amd64", "windows/amd64"}, OutDir: "dist/release/hello"}
		if _, err := Publish(root, protocol.PublishParams{Project: protocol.Project{Name: "hello", Root: "apps/hello"}, Version: "1.0.0", Assets: spec}); err != nil {
			t.Fatal(err)
		}
		sums, err := os.ReadFile(filepath.Join(root, "dist/release/hello/checksums.txt"))
		if err != nil {
			t.Fatal(err)
		}
		return string(sums)
	}
	if first, second := build(), build(); first != second {
		t.Fatalf("two builds differ:\n%s\n%s", first, second)
	}
}

// A native (cgo) app is built on each OS: only for the runner's own
// platform, without a checksums file one runner cannot complete.
func TestBuildAssetsOfANativeApp(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go unavailable")
	}
	root := t.TempDir()
	app := filepath.Join(root, "apps", "probe")
	_ = os.MkdirAll(app, 0o755)
	_ = os.WriteFile(filepath.Join(app, "go.mod"), []byte("// smith:native\nmodule example.com/probe\n\ngo 1.26.0\n"), 0o644)
	_ = os.WriteFile(filepath.Join(app, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644)
	host := runtime.GOOS + "/" + runtime.GOARCH
	project := protocol.Project{Name: "probe", Root: "apps/probe", Tags: []string{"native"}}
	params := protocol.PublishParams{Project: project, Version: "1.0.0", Assets: &protocol.AssetSpec{Platforms: []string{host}, OutDir: "dist/release/probe"}}
	res, err := Publish(root, params)
	if err != nil || len(res.Assets) != 1 || strings.Contains(res.Assets[0], "checksums") {
		t.Fatalf("%+v %v", res, err)
	}
	other := "linux/arm64"
	if host == other {
		other = "windows/amd64"
	}
	params.Assets.Platforms = []string{other}
	if _, err := Publish(root, params); err == nil || !strings.Contains(err.Error(), "native (cgo)") {
		t.Fatalf("a native app cannot be cross-compiled: %v", err)
	}
}
