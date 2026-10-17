package convertplan

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDeclarationSpecifiersNameTheFileEachSpecifierResolvesTo(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed")
	}
	root := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		file := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(DeclarationSpecifiersFile, DeclarationSpecifiers)
	write("dist/src/index.d.ts", `export { a } from './hook.contract';
export type { B } from "./hooks";
export * from './already.js';
export { c } from 'zod';
export { d } from './missing';
export type E = import('./hooks').E;
`)
	write("dist/src/hook.contract.d.ts", "export declare const a: number;\n")
	write("dist/src/already.d.ts", "export {};\n")
	write("dist/src/hooks/index.d.ts", "export type { B } from '../hook.contract';\nexport type E = string;\n")
	cmd := exec.Command(node, DeclarationSpecifiersFile, "dist/src")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	read := func(name string) string {
		raw, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		return string(raw)
	}
	want := `export { a } from './hook.contract.js';
export type { B } from "./hooks/index.js";
export * from './already.js';
export { c } from 'zod';
export { d } from './missing';
export type E = import('./hooks/index.js').E;
`
	if got := read("dist/src/index.d.ts"); got != want {
		t.Errorf("index.d.ts =\n%s\nwant\n%s", got, want)
	}
	if got := read("dist/src/hooks/index.d.ts"); got != "export type { B } from '../hook.contract.js';\nexport type E = string;\n" {
		t.Errorf("hooks/index.d.ts = %q", got)
	}
}
