package discovery

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPackageNames(t *testing.T) {
	src := `
import { a } from '@acme/ui'
import b from "lodash/fp"
import './side-effect'
import type { C } from '@acme/ui/types'
export * from '@acme/core'
const d = require('express')
const e = await import('zod')
import fs from 'node:fs'
import x from '../local'
`
	got := PackageNames(src)
	want := []string{"@acme/ui", "lodash", "@acme/ui", "@acme/core", "express", "zod"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestScanImportsSkipsBuildOutput(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("src/main.ts", "import { x } from '@acme/ui'")
	write("dist/main.js", "require('@acme/stale')")
	write("node_modules/y/index.js", "require('@acme/other')")
	write("src/types.d.ts", "import '@acme/declared'")
	got, err := ScanImports(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"@acme/ui"}) {
		t.Fatalf("got %v", got)
	}
}
