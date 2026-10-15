package convertplan

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/apps/mnci/nxworkspace"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func target(t *testing.T, raw string) nxworkspace.Target {
	t.Helper()
	var tg nxworkspace.Target
	if err := json.Unmarshal([]byte(raw), &tg); err != nil {
		t.Fatal(err)
	}
	return tg
}

func project(name, root, typ string, files []string, targets map[string]nxworkspace.Target, tags ...string) *nxworkspace.Project {
	p := &nxworkspace.Project{Name: name, Root: root, Type: typ, Targets: targets, Tags: tags}
	for _, f := range files {
		p.Files = append(p.Files, root+"/"+f)
	}
	return p
}

func findings(plan *Plan, severity Severity) string {
	var out []string
	for _, f := range plan.Findings {
		if f.Severity == severity {
			out = append(out, f.Project+" "+f.Subject+": "+f.Message)
		}
	}
	return strings.Join(out, "\n")
}

// legacyWorkspace looks like what the old mnci generates.
func legacyWorkspace(t *testing.T) (*nxworkspace.Workspace, Input) {
	var nx nxworkspace.NxJSON
	_ = json.Unmarshal([]byte(`{
		"plugins": ["@nx-go/nx-go"],
		"release": {"projects": ["packages/*", "python-packages/*", "tag:release:go", "!tag:type:go-lib"], "releaseTag": {"pattern": "{projectName}@{version}"}},
		"mnci": {"workspaceName": "legacy", "scope": "@legacy", "ci": "both"}
	}`), &nx)
	goLint := target(t, `{"executor": "@nx-go/nx-go:lint", "dependsOn": ["slice-check"], "options": {"linter": "golangci-lint", "args": ["run"]}}`)
	slice := target(t, `{"executor": "nx:run-commands", "cache": true, "inputs": ["{projectRoot}/**/*.go", "{workspaceRoot}/tools/go-slice-check.cjs"], "options": {"command": "node tools/go-slice-check.cjs apps/tool"}}`)
	tool := project("tool", "apps/tool", "app", []string{"go.mod", "main.go"}, map[string]nxworkspace.Target{
		"build":       target(t, `{"executor": "@nx-go/nx-go:build"}`),
		"lint":        goLint,
		"slice-check": slice,
		"build-all":   target(t, `{"executor": "nx:run-commands", "options": {"command": "node -e \"build()\""}}`),
		"tidy":        target(t, `{"executor": "@nx-go/nx-go:tidy"}`),
	}, "type:go-app", "release:go")
	tool.Release = map[string]json.RawMessage{"asset": json.RawMessage(`{"name": "{product}_{version}_{os}_{arch}.{ext}", "product": "tl"}`)}
	gokit := project("gokit", "packages/gokit", "lib", []string{"go.mod"}, map[string]nxworkspace.Target{"test": target(t, `{"executor": "@nx-go/nx-go:test"}`)}, "type:go-lib")
	kit := project("@legacy/kit", "packages/kit", "lib", []string{"package.json", "rollup.config.cjs"}, map[string]nxworkspace.Target{
		"build":      target(t, `{"executor": "nx:run-commands", "options": {"cwd": "packages/kit", "command": "rollup -c rollup.config.cjs"}}`),
		"test":       target(t, `{"executor": "nx:run-commands", "dependsOn": ["^build"], "options": {"cwd": "packages/kit", "command": "jest", "env": {"TS_NODE": "1"}, "passWithNoTests": true}}`),
		"watch-deps": target(t, `{"executor": "nx:run-commands", "options": {"command": "npx nx watch"}}`),
	})
	kit.Package = &nxworkspace.PackageJSON{Name: "@legacy/kit", Module: "./dist/index.esm.js", Types: "./dist/src/index.d.ts"}
	pykit := project("pykit", "python-packages/pykit", "lib", []string{"pyproject.toml"}, map[string]nxworkspace.Target{
		"typecheck": target(t, `{"executor": "@mnci/nx-python-pip:typecheck"}`),
		"start":     target(t, `{"executor": "nx:run-commands", "options": {"cwd": "python-packages/pykit", "command": "python3 main.py"}}`),
	})
	flutter := project("mobile", "apps/mobile", "app", []string{"pubspec.yaml", "project.json"}, map[string]nxworkspace.Target{})
	flutter.HasProjectJSON = true
	tool.HasProjectJSON = true
	w := &nxworkspace.Workspace{Root: "/ws", Nx: nx, NxVersion: "23.2.0", Projects: []*nxworkspace.Project{tool, gokit, kit, pykit, flutter}}
	w.RootPackage.Scripts = map[string]string{"build": "nx run-many -t build", "tool:qa": "nx run tool:lint", "python:install": "node -e x", "deploy": "nx deploy-it"}
	w.RootPackage.DevDependencies = map[string]string{"nx": "23.2.0", "@nx/js": "23.2.0", "@mnci/cli": "4", "eslint": "10"}
	files := []string{"nx.json", "package.json", "tools/go-app-release.cjs", "packages/kit/rollup.config.cjs", "jest.preset.js", "requirements-dev.txt", "apps/tool/go.mod", ".github/workflows/ci.yml"}
	contents := map[string]string{
		"packages/kit/rollup.config.cjs": "const { withNx } = require('@nx/rollup/with-nx')\nmodule.exports = withNx({ main: './src/index.ts', outputPath: './dist', tsConfig: './tsconfig.lib.json', format: ['esm'] })\n",
		"tools/go-app-release.cjs":       "require('nx/release')",
		"jest.preset.js":                 "const nxPreset = require('@nx/jest/preset').default\nmodule.exports = { ...nxPreset }\n",
		"requirements-dev.txt":           "build\nruff\npytest\npip-audit\ntwine\nmypy\n",
		"apps/tool/go.mod":               "module github.com/acme/legacy/apps/tool\n",
		".github/workflows/ci.yml":       "# mnci:slot after-install\n# mnci:slot-end after-install\n",
	}
	return w, Input{DirName: "legacy", Files: files, Tags: []string{"@legacy/kit@0.1.0"}, Read: func(f string) string { return contents[f] }}
}

func TestPlanLegacyWorkspace(t *testing.T) {
	w, in := legacyWorkspace(t)
	plan := MakePlan(w, in)

	if plan.Config.Workspace.Name != "legacy" || plan.Config.Workspace.Scope != "@legacy" || plan.Config.Workspace.ModulePrefix != "github.com/acme/legacy" {
		t.Fatalf("workspace %+v", plan.Config.Workspace)
	}
	if strings.Join(plan.Config.CI.Providers, ",") != "github,azure" || !plan.Config.Release.GitHubReleases {
		t.Fatalf("ci/release %+v %+v", plan.Config.CI, plan.Config.Release)
	}
	kinds := map[string]protocol.ProjectKind{}
	for _, p := range plan.Projects {
		kinds[p.Name] = p.Kind
	}
	if kinds["tool"] != protocol.ProjectApp || kinds["gokit"] != protocol.ProjectPackage || kinds["kit"] != protocol.ProjectPackage || kinds["pykit"] != protocol.ProjectPackage {
		t.Fatalf("kinds %v", kinds)
	}
	if !plan.Project("pykit").KindMarker || plan.Project("kit").KindMarker {
		t.Fatal("python-packages/ needs a kind marker; packages/ does not")
	}

	tool := plan.Config.Projects["tool"]
	if tool.Release == nil || tool.Release.AssetName != "{product}_{version}_{os}_{arch}.{ext}" || tool.Release.Product != "tl" {
		t.Fatalf("tool release %+v", tool.Release)
	}
	if lint := tool.Tasks["lint"]; strings.Join(lint.Command, " ") != "golangci-lint run" || lint.DependsOn[0] != "slice-check" {
		t.Fatalf("lint %+v", lint)
	}
	slice := tool.Tasks["slice-check"]
	if slice.Cwd != "." || strings.Join(slice.Inputs, ",") != "**/*.go,{workspaceRoot}/tools/go-slice-check.cjs" || slice.Cache == nil || !*slice.Cache {
		t.Fatalf("slice-check %+v", slice)
	}
	if _, kept := tool.Tasks["build-all"]; kept {
		t.Fatal("the old mnci's build-all is replaced by release assets")
	}
	if plan.Config.Projects["kit"].TagPrefix != "@legacy/kit@" {
		t.Fatalf("kit keeps its tags: %+v", plan.Config.Projects["kit"])
	}
	if plan.Config.Projects["gokit"].TagPrefix != "" {
		t.Fatal("Go libraries keep mnci's <dir>/v tags")
	}

	kit := plan.Project("kit")
	if !strings.HasPrefix(kit.Scripts["build"], "esbuild src/index.ts --bundle --packages=external --format=esm --outfile=dist/index.esm.js") || !strings.Contains(kit.Scripts["build"], "--declarationDir dist/src") {
		t.Fatalf("rollup build %q", kit.Scripts["build"])
	}
	if kit.Scripts["test"] != "jest --passWithNoTests" || plan.Config.Projects["kit"].Tasks["test"].Env["TS_NODE"] != "1" {
		t.Fatalf("test script %q %+v", kit.Scripts["test"], plan.Config.Projects["kit"].Tasks["test"])
	}
	if start := plan.Config.Projects["pykit"].Tasks["start"]; strings.Join(start.Command, " ") != "python main.py" {
		t.Fatalf("python3 becomes the venv's python: %v", start.Command)
	}

	if plan.RootScripts["build"] != "mnci run build" || plan.RootScripts["tool:qa"] != "" || plan.RootScripts["python:install"] != "" {
		t.Fatalf("root scripts %v", plan.RootScripts)
	}
	if _, ok := plan.RootScripts["deploy"]; ok {
		t.Fatal("an untranslatable script stays as it is")
	}
	if strings.Join(plan.RemoveDependencies, ",") != "@mnci/cli,@nx/js,nx" {
		t.Fatalf("removes %v", plan.RemoveDependencies)
	}
	for _, path := range []string{"nx.json", ".nx", "apps/tool/project.json", "tools/go-app-release.cjs", "packages/kit/rollup.config.cjs"} {
		if !contains(plan.Delete, path) {
			t.Errorf("must delete %s: %v", path, plan.Delete)
		}
	}
	if contains(plan.Delete, "apps/mobile/project.json") {
		t.Error("an unsupported project keeps its project.json")
	}
	if !strings.Contains(plan.Write["jest.preset.js"], "moduleNameMapper") || !strings.Contains(plan.Write["requirements-dev.txt"], "setuptools==") {
		t.Fatalf("writes %v", plan.WriteFiles)
	}

	manual := findings(plan, Manual)
	if !strings.Contains(manual, "script deploy") || strings.Contains(manual, "go-app-release") || strings.Contains(manual, "rollup") {
		t.Fatalf("manual findings:\n%s", manual)
	}
	if !strings.Contains(findings(plan, Unsupported), "mobile apps/mobile: Flutter/Dart project") {
		t.Fatalf("unsupported:\n%s", findings(plan, Unsupported))
	}
	if plan.Blocked(false) || !plan.Blocked(true) {
		t.Fatal("manual work blocks only --strict")
	}
}

func TestPlanBlocksDuplicateDirectoryNames(t *testing.T) {
	w := &nxworkspace.Workspace{Projects: []*nxworkspace.Project{
		project("a-ui", "apps/ui", "app", []string{"package.json"}, nil),
		project("l-ui", "libs/ui", "lib", []string{"package.json"}, nil),
	}}
	plan := MakePlan(w, Input{DirName: "x", Read: func(string) string { return "" }})
	if !plan.Blocked(false) || !strings.Contains(findings(plan, Blocker), "would both be called \"ui\"") {
		t.Fatalf("%v", plan.Findings)
	}
}

func TestPlanPlainNxReleasesOnlyPublicPackages(t *testing.T) {
	pub := project("@org/pub", "libs/pub", "lib", []string{"package.json"}, nil)
	pub.Package = &nxworkspace.PackageJSON{Name: "@org/pub"}
	priv := project("@org/priv", "libs/priv", "lib", []string{"package.json"}, nil)
	priv.Package = &nxworkspace.PackageJSON{Name: "@org/priv", Private: true}
	w := &nxworkspace.Workspace{Projects: []*nxworkspace.Project{pub, priv}}
	w.RootPackage.Name = "@org/source"
	plan := MakePlan(w, Input{DirName: "plain", Read: func(string) string { return "" }})
	if plan.Project("pub").Kind != protocol.ProjectPackage || !plan.Project("pub").KindMarker || plan.Project("priv").Kind != protocol.ProjectLib {
		t.Fatalf("%+v %+v", plan.Project("pub"), plan.Project("priv"))
	}
	if plan.Config.Workspace.Name != "plain" || plan.Config.Workspace.Scope != "@org" {
		t.Fatalf("%+v", plan.Config.Workspace)
	}
}
