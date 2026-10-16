package convertplan

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/MoNecromanCI/mnci/apps/mnci/nxworkspace"
)

// NxPackage reports whether a package belongs to Nx, so a converted
// workspace no longer needs it.
func NxPackage(name string) bool {
	for _, prefix := range []string{"@nx/", "@nrwl/", "@nx-go/", "@monodon/", "@nxlv/", "@mnci/nx-"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	switch name {
	case "nx", "nx-cloud", "@mnci/cli":
		return true
	}
	return false
}

// JestPreset replaces a jest.preset.js that re-exported @nx/jest/preset:
// Nx's preset values. Its resolver is replaced by the one thing a
// workspace without Nx needs from it: an ESM-style "./x.js" import of a
// TypeScript file resolves to "./x" (npm workspaces resolve the packages).
const JestPreset = `// Written by mnci convert: the values @nx/jest/preset provided, without Nx.
module.exports = {
  testMatch: ['**/?(*.)+(spec|test).?([mc])[jt]s?(x)'],
  moduleFileExtensions: ['ts', 'js', 'mts', 'mjs', 'cts', 'cjs', 'html'],
  coverageReporters: ['html'],
  moduleNameMapper: { '^(\\.{1,2}/.*)\\.js$': '$1' },
  transform: {
    '^.+\\.(ts|js|mts|mjs|cts|cjs|html)$': ['ts-jest', { tsconfig: '<rootDir>/tsconfig.spec.json' }],
  },
  testEnvironment: 'jsdom',
  modulePathIgnorePatterns: ['<rootDir>/dist/', '<rootDir>/out-tsc/'],
  testEnvironmentOptions: {
    customExportConditions: ['node', 'require', 'default'],
  },
}
`

// configFile reports whether a file is configuration or tooling that may
// import Nx at run time.
func configFile(f string) bool {
	base := path.Base(f)
	for _, skip := range []string{"node_modules", ".nx", "testdata", "fixtures", "__fixtures__"} {
		if strings.HasPrefix(f, skip+"/") || strings.Contains(f, "/"+skip+"/") {
			return false
		}
	}
	for _, ext := range []string{".js", ".cjs", ".mjs", ".ts", ".cts", ".mts"} {
		if !strings.HasSuffix(base, ext) {
			continue
		}
		stem := strings.TrimSuffix(base, ext)
		if strings.HasSuffix(stem, ".config") || strings.HasPrefix(stem, "jest.preset") || strings.HasPrefix(f, "tools/") {
			return true
		}
	}
	return strings.HasPrefix(base, ".eslintrc")
}

// planNxRemoval schedules every Nx file and dependency for removal, rewrites
// the configuration that only re-exported Nx, and reports the rest.
func planNxRemoval(plan *Plan, w *nxworkspace.Workspace, files []string, read func(string) string) {
	plan.DeletePath("nx.json")
	plan.DeletePath(".nx")
	for _, p := range w.Projects {
		if pp := plan.ProjectAt(p.Root); p.HasProjectJSON && pp != nil && pp.Supported {
			plan.DeletePath(p.Root + "/project.json")
		}
	}
	for _, f := range files {
		if f == "project.json" {
			plan.DeletePath(f)
		}
	}
	var remove []string
	for _, deps := range []map[string]string{w.RootPackage.DevDependencies, w.RootPackage.Dependencies} {
		for name := range deps {
			if NxPackage(name) {
				remove = append(remove, name)
			}
		}
	}
	sort.Strings(remove)
	plan.RemoveDependencies = remove
	for name := range w.RootPackage.Overrides {
		if NxPackage(name) {
			plan.RemoveOverrides = append(plan.RemoveOverrides, name)
		}
	}
	sort.Strings(plan.RemoveOverrides)
	if len(remove) > 0 {
		plan.Add(Mapped, "", "package.json", "removes "+strings.Join(remove, ", "))
	}

	for _, f := range files {
		switch f {
		case "pnpm-lock.yaml", "pnpm-workspace.yaml":
			plan.Add(Unsupported, "", f, "pnpm workspaces: the mnci Node plugin runs npm; switch to npm workspaces (npm install creates package-lock.json)")
		case "yarn.lock":
			plan.Add(Unsupported, "", f, "yarn workspaces: the mnci Node plugin runs npm; switch to npm workspaces (npm install creates package-lock.json)")
		}
		if !configFile(f) {
			continue
		}
		text := read(f)
		if !strings.Contains(text, "@nx/") && !strings.Contains(text, "'nx/") && !strings.Contains(text, "\"nx/") && !strings.Contains(text, "@nrwl/") {
			continue
		}
		base := path.Base(f)
		switch {
		case path.Dir(f) == "." && strings.HasPrefix(base, "jest.preset.") && strings.Contains(text, "@nx/jest/preset"):
			plan.WriteFile(f, JestPreset)
			plan.Add(Changed, "", f, "rewritten without @nx/jest: the same preset values, written out")
		case path.Dir(f) == "." && strings.HasPrefix(base, "jest.config.") && strings.Contains(text, "getJestProjects"):
			plan.DeletePath(f)
			plan.Add(Dropped, "", f, "listed every Jest project for Nx; mnci runs each project's tests itself")
		case strings.HasPrefix(base, "rollup.config.") && strings.Contains(text, "with-nx"):
			if !replaceRollup(plan, w, f, text) {
				plan.Add(Manual, "", f, "builds with @nx/rollup's withNx, which needs Nx; rewrite the build without it")
			}
		case strings.HasPrefix(base, "eslint.config."):
			rewritten, ok := RewriteNxESLint(text)
			if !ok {
				plan.Add(Manual, "", f, "uses @nx/eslint-plugin in a way convert cannot rewrite; replace it with typescript-eslint")
				break
			}
			plan.WriteFile(f, rewritten)
			if strings.Contains(rewritten, "typescript-eslint") {
				if plan.AddDevDependencies == nil {
					plan.AddDevDependencies = map[string]string{}
				}
				plan.AddDevDependencies["@eslint/js"] = ""
				plan.AddDevDependencies["typescript-eslint"] = ""
			}
			if strings.Contains(rewritten, "typescript-eslint") {
				plan.Add(Changed, "", f, "rewritten without @nx/eslint-plugin: its presets become @eslint/js and typescript-eslint, and the module-boundary rule (it reads the Nx graph) is removed")
			} else {
				plan.Add(Changed, "", f, "the @nx/dependency-checks rule (it reads the Nx graph) is removed")
			}
		case pathDeleted(plan, f):
		default:
			plan.Add(Manual, "", f, "imports Nx ("+nxImport(text)+"), which convert removes; rewrite it without Nx")
		}
	}
	if w.Nx.NxCloudID != "" || w.Nx.NxCloudToken != "" {
		plan.Add(Changed, "", "nx.json", "Nx Cloud is not used: mnci caches locally (.mnci/cache)")
	}
}

var (
	rollupMain     = regexp.MustCompile(`main:\s*['"]([^'"]+)['"]`)
	rollupFormat   = regexp.MustCompile(`format:\s*\[\s*['"](\w+)['"]`)
	rollupTSConfig = regexp.MustCompile(`tsConfig:\s*['"]([^'"]+)['"]`)
)

// replaceRollup swaps a library's @nx/rollup build for esbuild (the bundle)
// plus tsc (the declarations), writing to the paths its package.json
// already publishes. It reports false when those paths cannot be read.
func replaceRollup(plan *Plan, w *nxworkspace.Workspace, file, text string) bool {
	root := path.Dir(file)
	var project *nxworkspace.Project
	for _, p := range w.Projects {
		if p.Root == root {
			project = p
		}
	}
	pp := plan.ProjectAt(root)
	if project == nil || project.Package == nil || pp == nil || pp.Language != "node" {
		return false
	}
	main := rollupMain.FindStringSubmatch(text)
	tsconfig := rollupTSConfig.FindStringSubmatch(text)
	out := strings.TrimPrefix(project.Package.Module, "./")
	if out == "" {
		out = strings.TrimPrefix(project.Package.Main, "./")
	}
	types := strings.TrimPrefix(project.Package.Types, "./")
	if main == nil || tsconfig == nil || out == "" || types == "" {
		return false
	}
	format := "esm"
	if m := rollupFormat.FindStringSubmatch(text); m != nil {
		format = m[1]
	}
	entry := strings.TrimPrefix(main[1], "./")
	// The declarations keep their own build info: sharing the project's
	// would make `tsc --build` (typecheck) think the library is built when
	// its own outputs are missing.
	declarations := path.Dir(types)
	build := fmt.Sprintf("esbuild %s --bundle --packages=external --format=%s --outfile=%s --sourcemap && tsc -p %s --emitDeclarationOnly --declaration --declarationDir %s --tsBuildInfoFile %s/tsconfig.lib.tsbuildinfo",
		entry, format, out, strings.TrimPrefix(tsconfig[1], "./"), declarations, declarations)
	if pp.Scripts == nil {
		pp.Scripts = map[string]string{}
	}
	pp.Scripts["build"] = build
	if plan.AddDevDependencies == nil {
		plan.AddDevDependencies = map[string]string{}
	}
	plan.AddDevDependencies["esbuild"] = ""
	plan.DeletePath(file)
	plan.Add(Changed, pp.Name, "build", "Rollup with @nx/rollup becomes esbuild plus tsc declarations, into the same "+out+" and "+types+": "+build)
	return true
}

func pathDeleted(plan *Plan, f string) bool {
	for _, d := range plan.Delete {
		if d == f || strings.HasPrefix(f, d+"/") {
			return true
		}
	}
	return false
}

// nxImport names the first Nx module a file imports.
func nxImport(text string) string {
	for _, marker := range []string{"@nx/", "@nrwl/", "'nx/", "\"nx/"} {
		i := strings.Index(text, marker)
		if i < 0 {
			continue
		}
		start := i
		if marker[0] == '\'' || marker[0] == '"' {
			start++
		}
		end := strings.IndexAny(text[start:], "'\"` \n")
		if end < 0 {
			end = len(text) - start
		}
		return text[start : start+end]
	}
	return "nx"
}
