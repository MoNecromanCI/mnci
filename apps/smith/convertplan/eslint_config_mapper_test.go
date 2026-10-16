package convertplan

import (
	"os"
	"strings"
	"testing"
)

func TestRewriteNxESLintRoot(t *testing.T) {
	src, _ := os.ReadFile("testdata/nx-root-eslint.config.mjs")
	got, ok := RewriteNxESLint(string(src))
	if !ok {
		t.Fatalf("not rewritten:\n%s", got)
	}
	for _, want := range []string{"import js from '@eslint/js';", "import tseslint from 'typescript-eslint';", "  js.configs.recommended,", "  ...tseslint.configs.recommended,", "ignores: ['**/dist'", "// Override or add rules here"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "nx") || strings.Contains(got, "enforce-module-boundaries") {
		t.Fatalf("Nx left in:\n%s", got)
	}
}

func TestRewriteNxESLintProject(t *testing.T) {
	src, _ := os.ReadFile("testdata/nx-project-eslint.config.mjs")
	got, ok := RewriteNxESLint(string(src))
	if !ok {
		t.Fatalf("not rewritten:\n%s", got)
	}
	want := `import baseConfig from '../../eslint.config.mjs';

export default [
  ...baseConfig,
  {
    ignores: ['**/out-tsc'],
  },
];
`
	if strings.ReplaceAll(got, "\r\n", "\n") != want {
		t.Fatalf("got:\n%s", got)
	}
}

func TestRewriteNxESLintRefusesUnknownPresets(t *testing.T) {
	src := "import nx from '@nx/eslint-plugin';\nexport default [...nx.configs['flat/react']];\n"
	if _, ok := RewriteNxESLint(src); ok {
		t.Fatal("a preset with no plain equivalent must stay manual")
	}
}
