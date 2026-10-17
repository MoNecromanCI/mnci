package cipipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// runtime is a toolchain whose plugin gives both providers' steps.
func runtime(name, version, action, task, input string) protocol.Toolchain {
	return protocol.Toolchain{
		Name: name, Version: version,
		GitHub: &protocol.ActionStep{Uses: action, Inputs: []protocol.Input{{Name: name + "-version", Value: version}, {Name: "cache", Value: "false"}}},
		Azure:  &protocol.ActionStep{Uses: task, Inputs: []protocol.Input{{Name: input, Value: version + ".x"}}},
	}
}

func model() Model {
	var c Contributions
	c.Add(protocol.CIResult{Toolchains: []protocol.Toolchain{runtime("python", "3.12", "actions/setup-python@v5", "UsePythonVersion@0", "versionSpec")}, Credentials: []protocol.Credential{{Env: "PYPI_TOKEN", Required: true}}})
	c.Add(protocol.CIResult{Toolchains: []protocol.Toolchain{runtime("node", "24", "actions/setup-node@v4", "UseNode@1", "version"), {Name: "python", Version: "9"}}, Credentials: []protocol.Credential{{Env: "NODE_AUTH_TOKEN"}}})
	c.Add(protocol.CIResult{Toolchains: []protocol.Toolchain{runtime("go", "1.26", "actions/setup-go@v5", "GoTool@0", "version")}})
	// A plugin for a language smith does not know: an install command only,
	// or nothing at all.
	c.Add(protocol.CIResult{Toolchains: []protocol.Toolchain{{Name: "zig", Version: "0.14", Install: "curl -fsSL https://example.com/zig.sh | sh\nzig version"}, {Name: "rust", Version: "1.90"}}})
	return NewModel(c, "dev", "main")
}

// An extension package's plugin declares VSCE_PAT; both providers pass it
// to the release step, where vsce publish reads it.
func TestTheReleaseStepGetsTheMarketplaceToken(t *testing.T) {
	var c Contributions
	c.Add(protocol.CIResult{Toolchains: []protocol.Toolchain{runtime("node", "24", "actions/setup-node@v4", "UseNode@1", "version")}, Credentials: []protocol.Credential{{Env: "VSCE_PAT", Required: true}}})
	m := NewModel(c, "dev", "main")
	gh := RenderGitHub(m)
	if release := gh[strings.Index(gh, "      - name: Release\n"):]; !strings.Contains(release, "          VSCE_PAT: ${{ secrets.VSCE_PAT }}\n") {
		t.Errorf("GitHub release step:\n%s", release)
	}
	az := RenderAzure(m)
	if release := az[strings.Index(az, "smith ci release"):]; !strings.Contains(release, "VSCE_PAT: $(VSCE_PAT)\n") {
		t.Errorf("Azure release step:\n%s", release)
	}
}

func TestRenderersCarryTheSameSteps(t *testing.T) {
	m := model()
	gh, az := RenderGitHub(m), RenderAzure(m)
	for _, want := range []string{
		"      - uses: actions/setup-node@v4\n        with:\n          node-version: '24'\n          cache: false\n",
		"python-version: '3.12'", "go-version: '1.26'",
		"      - name: Install zig\n        run: |\n          curl -fsSL https://example.com/zig.sh | sh\n          zig version\n",
		"      # smith: no plugin says how to install rust 1.90",
		"smith ci setup", "smith ci audit", "smith ci verify", "smith ci release", "smith release --dry-run", "NODE_AUTH_TOKEN: ${{ secrets.NODE_AUTH_TOKEN }}", "fetch-depth: 0", "SMITH_VERSION: latest", "github.ref_name == 'main'"} {
		if !strings.Contains(gh, want) {
			t.Errorf("GitHub workflow lacks %q", want)
		}
	}
	for _, want := range []string{
		"  - task: UseNode@1\n    inputs:\n      version: '24.x'\n",
		"versionSpec: '3.12.x'", "  - task: GoTool@0",
		"  - bash: |\n      curl -fsSL https://example.com/zig.sh | sh\n      zig version\n    displayName: Install zig\n",
		"  # smith: no plugin says how to install rust 1.90",
		"smith ci setup", "smith ci audit", "smith ci verify", "smith ci release", "smith release --dry-run", "PYPI_TOKEN: $(PYPI_TOKEN)", "fetchDepth: 0", "persistCredentials: true", "refs/heads/main"} {
		if !strings.Contains(az, want) {
			t.Errorf("Azure pipeline lacks %q", want)
		}
	}
	// The phase steps are the same in both providers: the lockstep rule.
	phases := func(s string) []string {
		var out []string
		for _, line := range strings.Split(s, "\n") {
			if i := strings.Index(line, "smith ci "); i >= 0 {
				out = append(out, strings.TrimSpace(line[i:]))
			}
		}
		return out
	}
	if strings.Join(phases(gh), ",") != strings.Join(phases(az), ",") {
		t.Fatalf("providers drifted:\n%v\n%v", phases(gh), phases(az))
	}
	if RenderGitHub(m) != gh {
		t.Fatal("rendering must be deterministic")
	}
}

func TestNoToolchainsWhenNoProjects(t *testing.T) {
	gh := RenderGitHub(NewModel(Contributions{}, "1.2.3", "main"))
	if strings.Contains(gh, "setup-node") || !strings.Contains(gh, "SMITH_VERSION: 1.2.3") || strings.Contains(gh, "env:\n          NODE") {
		t.Fatal(gh)
	}
}

func TestWriteKeepsSlotsAndReportsDrift(t *testing.T) {
	root := t.TempDir()
	cfg := workspace.Config{CI: workspace.CISection{Providers: []string{"github", "azure"}}}
	m := model()
	changed, _, err := Write(root, cfg, m)
	if err != nil || len(changed) != 2 {
		t.Fatalf("changed %v err %v", changed, err)
	}
	if drift := Drift(root, cfg, m); len(drift) != 0 {
		t.Fatalf("fresh pipelines drift: %v", drift)
	}
	path := filepath.Join(root, ".github", "workflows", "ci.yml")
	raw, _ := os.ReadFile(path)
	custom := strings.Replace(string(raw), "      # smith:slot after-verify\n", "      # smith:slot after-verify\n      - name: Team step\n        run: make e2e\n", 1)
	_ = os.WriteFile(path, []byte(custom), 0o644)
	if drift := Drift(root, cfg, m); len(drift) != 0 {
		t.Fatalf("slot content is not drift: %v", drift)
	}
	m.Toolchains = m.Toolchains[:1]
	if drift := Drift(root, cfg, m); len(drift) != 2 {
		t.Fatalf("changed toolchains must drift: %v", drift)
	}
	if _, _, err := Write(root, cfg, m); err != nil {
		t.Fatal(err)
	}
	raw, _ = os.ReadFile(path)
	if !strings.Contains(string(raw), "run: make e2e") || strings.Contains(string(raw), "setup-node") {
		t.Fatalf("regeneration must keep the slot and drop node:\n%s", raw)
	}
	if changed, _, _ := Write(root, cfg, m); len(changed) != 0 {
		t.Fatalf("an unchanged workspace rewrites nothing: %v", changed)
	}
}

func TestThePreviewIsPostedOnPullRequests(t *testing.T) {
	m := model()
	gh, az := RenderGitHub(m), RenderAzure(m)
	for _, want := range []string{"  pull-requests: write\n", "run: smith release --dry-run --comment\n        env:\n          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}\n"} {
		if !strings.Contains(gh, want) {
			t.Errorf("GitHub workflow lacks %q", want)
		}
	}
	if !strings.Contains(az, "bash: smith release --dry-run --comment\n    displayName: Release preview\n    condition: and(succeeded(), eq(variables['Build.Reason'], 'PullRequest'))\n    env:\n      SYSTEM_ACCESSTOKEN: $(System.AccessToken)\n") {
		t.Errorf("Azure pipeline lacks the preview comment step:\n%s", az)
	}
}

func TestPackOnlyWhenAProjectPackages(t *testing.T) {
	m := model()
	if strings.Contains(RenderGitHub(m), "smith ci pack") || strings.Contains(RenderAzure(m), "smith ci pack") {
		t.Fatal("no project packages: no pack step")
	}
	m.Pack = true
	gh, az := RenderGitHub(m), RenderAzure(m)
	for _, want := range []string{"run: smith ci pack", "uses: actions/upload-artifact@v4", "name: drop", "path: dist/drop", "if-no-files-found: ignore"} {
		if !strings.Contains(gh, want) {
			t.Errorf("GitHub workflow lacks %q", want)
		}
	}
	for _, want := range []string{"bash: smith ci pack", "publish: dist/drop", "artifact: drop", "refs/heads/main"} {
		if !strings.Contains(az, want) {
			t.Errorf("Azure pipeline lacks %q", want)
		}
	}
	if strings.Index(gh, "smith ci pack") > strings.Index(gh, "smith ci release") {
		t.Error("pack runs before release")
	}
}

func TestGitHubWorkflowMayRequestAnOIDCTokenForNPMProvenance(t *testing.T) {
	m := model()
	if strings.Contains(RenderGitHub(m), "id-token") {
		t.Error("id-token: write without an npm-published app")
	}
	m.Provenance = true
	if !strings.Contains(RenderGitHub(m), "  pull-requests: write\n  id-token: write\n\nconcurrency:") {
		t.Errorf("no id-token permission:\n%s", RenderGitHub(m))
	}
}

func TestGitHubWorkflowGrantsThePermissionsPluginsAskFor(t *testing.T) {
	var c Contributions
	c.Add(protocol.CIResult{Permissions: []string{"packages: write"}})
	c.Add(protocol.CIResult{Permissions: []string{"packages: write"}})
	m := NewModel(c, "1.0.0", "main")
	if gh := RenderGitHub(m); !strings.Contains(gh, "  pull-requests: write\n  packages: write\n\nconcurrency:") {
		t.Errorf("packages: write once, with the other permissions:\n%s", gh)
	}
	if strings.Contains(RenderAzure(m), "packages") {
		t.Error("Azure Pipelines has no token permissions")
	}
}

func TestPubWorkflowOnlyWhenAnAppPublishesToPubDev(t *testing.T) {
	root := t.TempDir()
	github := workspace.Config{CI: workspace.CISection{Providers: []string{"github"}}}
	m := model()
	if _, ok := Rendered(github, m)[GitHubPubPath]; ok || strings.Contains(RenderGitHub(m), "actions: write") {
		t.Fatal("a pub workflow without a pub.dev app")
	}
	m.Pub = true
	files := Rendered(github, m)
	pub, ok := files[GitHubPubPath]
	if !ok {
		t.Fatal("no pub workflow")
	}
	for _, want := range []string{
		"on:\n  workflow_dispatch:\n    inputs:\n      project:",
		"permissions:\n  contents: read\n  id-token: write\n",
		"uses: dart-lang/setup-dart@v1",
		"SMITH_VERSION: latest",
		`run: smith release pub --project "$PROJECT" --version "$VERSION"`,
		"PROJECT: ${{ inputs.project }}",
		"# smith:slot before-publish",
	} {
		if !strings.Contains(pub, want) {
			t.Errorf("pub.yml lacks %q:\n%s", want, pub)
		}
	}
	if !strings.Contains(files[GitHubPath], "  pull-requests: write\n  actions: write\n") {
		t.Errorf("ci.yml cannot start the pub workflow:\n%s", files[GitHubPath])
	}
	azure := workspace.Config{CI: workspace.CISection{Providers: []string{"azure"}}}
	if _, ok := Rendered(azure, m)[GitHubPubPath]; ok {
		t.Error("a GitHub workflow for an Azure-only workspace")
	}

	if _, _, err := Write(root, github, m); err != nil {
		t.Fatal(err)
	}
	m.Pub = false
	if changed, _, err := Write(root, github, m); err != nil || !strings.Contains(strings.Join(changed, ","), GitHubPubPath) {
		t.Fatalf("dropping pub: changed %v %v", changed, err)
	}
	if _, err := os.Stat(filepath.Join(root, GitHubPubPath)); !os.IsNotExist(err) {
		t.Error("the pub workflow outlives the pub.dev app")
	}
}
