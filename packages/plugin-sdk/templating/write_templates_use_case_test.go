package templating

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWriteRendersPathsAndBodies(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"[[.Dir]]/main.go":    "package main // [[.Name]]",
		"[[.Dir]]/ci.yml":     "run: ${{ secrets.TOKEN }}",
		"[[.Dir]]/sub/doc.go": "package sub",
	}
	got, err := Write(root, files, map[string]string{"Dir": "apps/web", "Name": "web"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"apps/web/ci.yml", "apps/web/main.go", "apps/web/sub/doc.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	body, _ := os.ReadFile(filepath.Join(root, "apps", "web", "main.go"))
	if string(body) != "package main // web" {
		t.Fatalf("body %q", body)
	}
	body, _ = os.ReadFile(filepath.Join(root, "apps", "web", "ci.yml"))
	if string(body) != "run: ${{ secrets.TOKEN }}" {
		t.Fatalf("GitHub syntax must survive: %q", body)
	}
}

func TestWriteRefusesToOverwriteAndWritesNothing(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "b.txt"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Write(root, map[string]string{"a.txt": "x", "b.txt": "y"}, nil)
	if !errors.Is(err, ErrExists) {
		t.Fatalf("want ErrExists, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "a.txt")); err == nil {
		t.Fatal("a.txt must not be written when another target exists")
	}
}

func TestWriteFailsOnMissingKey(t *testing.T) {
	if _, err := Write(t.TempDir(), map[string]string{"a": "[[.Missing]]"}, map[string]string{}); err == nil {
		t.Fatal("expected a missing key error")
	}
}
