package layout

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tree(t *testing.T, files ...string) string {
	t.Helper()
	root := t.TempDir()
	for _, f := range files {
		p := filepath.Join(root, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestCheckGo(t *testing.T) {
	root := tree(t,
		"go.mod", "main.go", "doc.go", "main_windows.go",
		"greeting/doc.go", "greeting/greet_use_case.go", "greeting/greet_use_case_test.go", "greeting/run_process_use_case_windows.go",
		"greeting/stuff.go", "server.go",
		"utils/x_mapper.go",
		"testdata/whatever.go", ".hidden/bad.go", "vendor/bad.go",
	)
	r, err := Check(root, "go")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(r.Problems, "\n")
	for _, want := range []string{"greeting/stuff.go - a Go file is", "server.go - only doc.go and main.go", `utils/ - "utils" is a junk drawer`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if len(r.Problems) != 3 {
		t.Errorf("want 3 problems, got:\n%s", got)
	}
}

func TestCheckPython(t *testing.T) {
	root := tree(t,
		"pyproject.toml", "setup.py",
		"src/kit/__init__.py", "src/kit/__main__.py", "src/kit/cli.py",
		"src/kit/greeting/__init__.py", "src/kit/greeting/greet_use_case.py", "src/kit/greeting/test_greet_use_case.py", "src/kit/greeting/stuff.py",
		"src/kit/greeting/__main__.py",
		"src/kit/BadName/__init__.py", "src/kit/helpers/__init__.py",
		"tests/conftest.py", "tests/test_greet_use_case.py", "tests/fixtures/sample.py",
		".venv/lib/x.py", "src/kit.egg-info/x.py",
	)
	r, err := Check(root, "python")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(r.Problems, "\n")
	for _, want := range []string{
		"src/kit/cli.py - only __init__.py and __main__.py",
		"src/kit/greeting/stuff.py - a Python module is",
		"src/kit/greeting/__main__.py - a Python module is",
		"src/kit/BadName/ - a Python package is snake_case",
		`src/kit/helpers/ - "helpers" is a junk drawer`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if len(r.Problems) != 5 {
		t.Errorf("want 5 problems, got:\n%s", got)
	}
}

func TestCheckDart(t *testing.T) {
	root := tree(t,
		"pubspec.yaml",
		"lib/kit.dart", "lib/main.dart",
		"lib/src/greeting/greeting.dart", "lib/src/greeting/greet_use_case.dart", "lib/src/greeting/greeting_model.g.dart",
		"lib/src/greeting/stuff.dart", "lib/src/loose.dart", "lib/src/Orders/place_order_use_case.dart", "lib/src/helpers/x_mapper.dart",
		"bin/main.dart",
		"test/src/greeting/greet_use_case_test.dart", "test/widget_test.dart", "test/testing/fake_clock.dart", "test/sample.dart",
		"integration_test/app_test.dart",
		"android/app/src/main/kotlin/Main.dart", "example/lib/whatever.dart", ".dart_tool/x.dart", "build/x.dart",
	)
	r, err := Check(root, "dart")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(r.Problems, "\n")
	for _, want := range []string{
		"lib/src/greeting/stuff.dart - a Dart file in a slice is <name>_<role>.dart",
		"lib/src/loose.dart - lib/src holds slice folders",
		"lib/src/Orders/ - a Dart folder is lowercase_with_underscores",
		`lib/src/helpers/ - "helpers" is a junk drawer`,
		"test/sample.dart - a test is <basename>_test.dart",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if len(r.Problems) != 5 {
		t.Errorf("want 5 problems, got:\n%s", got)
	}
}

func TestCheckRust(t *testing.T) {
	root := tree(t,
		"Cargo.toml", "build.rs",
		"src/main.rs", "src/lib.rs", "src/loose.rs",
		"src/greeting/mod.rs", "src/greeting/greet_use_case.rs", "src/greeting/stuff.rs",
		"src/orders.rs", "src/orders/place_order_use_case.rs", "src/BadName/mod.rs",
		"src/bin/admin_tool.rs",
		"tests/greet_test.rs", "tests/testing/fake_clock.rs", "tests/smoke.rs",
		"benches/x.rs", "target/debug/x.rs",
	)
	r, err := Check(root, "rust")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(r.Problems, "\n")
	for _, want := range []string{
		"src/loose.rs - src/ holds main.rs, lib.rs and slice modules",
		"src/greeting/stuff.rs - a Rust file in a slice is <name>_<role>.rs",
		"src/BadName/ - a Rust module folder is snake_case",
		"tests/smoke.rs - an integration test is <name>_test.rs",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if len(r.Problems) != 4 {
		t.Errorf("want 4 problems, got:\n%s", got)
	}
}

func TestCheckWarnsAboutLargeSlices(t *testing.T) {
	var files []string
	for i := range SliceSize + 1 {
		files = append(files, fmt.Sprintf("big/step%c_use_case.go", 'a'+i), fmt.Sprintf("big/step%c_use_case_test.go", 'a'+i))
	}
	r, err := Check(tree(t, append(files, "big/doc.go", "small/one_model.go")...), "go")
	if err != nil {
		t.Fatal(err)
	}
	if !r.OK() || len(r.Warnings) != 1 || !strings.HasPrefix(r.Warnings[0], "big/ - 13 production files") {
		t.Fatalf("%+v", r)
	}
}

func TestCheckRejectsOtherLanguages(t *testing.T) {
	if _, err := Check(t.TempDir(), "node"); err == nil {
		t.Fatal("node is checked by @mnci/eslint-config, not here")
	}
}
