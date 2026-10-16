package npmbinaries

import (
	"archive/zip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func archive(t *testing.T, root, name, file string) string {
	t.Helper()
	f, err := os.Create(filepath.Join(root, name))
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, _ := zw.Create(file)
	_, _ = w.Write([]byte("binary " + file))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return name
}

func release(t *testing.T, root string) Release {
	return Release{
		Names: []string{"@acme/cli", "acme"}, Binary: "acme", Version: "5.0.0", OutDir: "out",
		Archives: []Archive{
			{Path: archive(t, root, "acme_windows_amd64.zip", "acme.exe"), OS: "windows", Arch: "amd64"},
			{Path: archive(t, root, "acme_linux_amd64.zip", "acme"), OS: "linux", Arch: "amd64"},
			{Path: archive(t, root, "acme_plan9_amd64.zip", "acme"), OS: "plan9", Arch: "amd64"},
		},
	}
}

// fakeNPM replaces the npm CLI for one test.
func fakeNPM(t *testing.T, f func(dir string, args ...string) (string, string, error)) {
	t.Helper()
	original := npm
	npm = f
	t.Cleanup(func() { npm = original })
}

func TestPublishSendsPlatformPackagesBeforeWrappersAndSkipsWhatTheRegistryHas(t *testing.T) {
	root := t.TempDir()
	t.Setenv("NODE_AUTH_TOKEN", "secret")
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "")
	var published []string
	fakeNPM(t, func(dir string, args ...string) (string, string, error) {
		if args[0] == "view" {
			if args[1] == "@acme/cli-linux-x64@5.0.0" {
				return "5.0.0\n", "", nil
			}
			return "", "", nil
		}
		raw, _ := os.ReadFile(filepath.Join(dir, "package.json"))
		published = append(published, filepath.Base(dir)+" "+strings.Join(args, " "))
		if !strings.Contains(string(raw), `"version": "5.0.0"`) {
			t.Errorf("%s package.json has no version", dir)
		}
		return "", "", nil
	})

	res, err := Publish(root, release(t, root), false)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"acme-cli-win32-x64 publish --access public",
		"acme-cli publish --access public",
		"acme publish --access public",
	}
	if !slices.Equal(published, want) {
		t.Errorf("published\n%v\nwant\n%v", published, want)
	}
	if !slices.Equal(res.Existing, []string{"@acme/cli-linux-x64@5.0.0"}) || !slices.Equal(res.Skipped, []string{"plan9/amd64"}) {
		t.Errorf("existing %v skipped %v", res.Existing, res.Skipped)
	}
	bin, err := os.ReadFile(filepath.Join(root, "out", "acme-cli-win32-x64", "bin", "acme.exe"))
	if err != nil || string(bin) != "binary acme.exe" {
		t.Errorf("platform binary = %q, %v", bin, err)
	}
	if _, err := os.Stat(filepath.Join(root, "out", "acme", "bin", "acme.js")); err != nil {
		t.Error("the wrapper has no shim")
	}
	if rc, _ := os.ReadFile(filepath.Join(root, "out", "acme", ".npmrc")); !strings.Contains(string(rc), "${NODE_AUTH_TOKEN}") {
		t.Errorf(".npmrc = %q", rc)
	}
}

func TestPublishAddsProvenanceAndTagsPrereleases(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "https://token")
	var args []string
	fakeNPM(t, func(_ string, a ...string) (string, string, error) {
		if a[0] == "publish" {
			args = a
		}
		return "", "", nil
	})
	r := release(t, root)
	r.Version, r.Repository = "5.0.0-rc.1", "https://github.com/acme/cli"
	if _, err := Publish(root, r, false); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(args, " "); got != "publish --access public --tag next --provenance" {
		t.Errorf("args = %s", got)
	}
}

func TestDryRunNamesThePackagesWithoutWriting(t *testing.T) {
	root := t.TempDir()
	fakeNPM(t, func(string, ...string) (string, string, error) {
		t.Error("npm ran on a dry run")
		return "", "", nil
	})
	res, err := Publish(root, release(t, root), true)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"@acme/cli-linux-x64@5.0.0", "@acme/cli-win32-x64@5.0.0", "@acme/cli@5.0.0", "acme@5.0.0"}
	if !slices.Equal(res.Published, want) {
		t.Errorf("published %v", res.Published)
	}
	if _, err := os.Stat(filepath.Join(root, "out")); !os.IsNotExist(err) {
		t.Error("a dry run wrote packages")
	}
}
