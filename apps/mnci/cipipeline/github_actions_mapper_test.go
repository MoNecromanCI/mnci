package cipipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/apps/mnci/workspace"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
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
	// A plugin for a language mnci does not know: an install command only,
	// or nothing at all.
	c.Add(protocol.CIResult{Toolchains: []protocol.Toolchain{{Name: "zig", Version: "0.14", Install: "curl -fsSL https://example.com/zig.sh | sh\nzig version"}, {Name: "rust", Version: "1.90"}}})
	return NewModel(c, "dev", "main")
}

func TestRenderersCarryTheSameSteps(t *testing.T) {
	m := model()
	gh, az := RenderGitHub(m), RenderAzure(m)
	for _, want := range []string{
		"      - uses: actions/setup-node@v4\n        with:\n          node-version: '24'\n          cache: false\n",
		"python-version: '3.12'", "go-version: '1.26'",
		"      - name: Install zig\n        run: |\n          curl -fsSL https://example.com/zig.sh | sh\n          zig version\n",
		"      # mnci: no plugin says how to install rust 1.90",
		"mnci ci setup", "mnci ci audit", "mnci ci verify", "mnci ci release", "mnci release --dry-run", "NODE_AUTH_TOKEN: ${{ secrets.NODE_AUTH_TOKEN }}", "fetch-depth: 0", "MNCI_VERSION: latest", "github.ref_name == 'main'"} {
		if !strings.Contains(gh, want) {
			t.Errorf("GitHub workflow lacks %q", want)
		}
	}
	for _, want := range []string{
		"  - task: UseNode@1\n    inputs:\n      version: '24.x'\n",
		"versionSpec: '3.12.x'", "  - task: GoTool@0",
		"  - bash: |\n      curl -fsSL https://example.com/zig.sh | sh\n      zig version\n    displayName: Install zig\n",
		"  # mnci: no plugin says how to install rust 1.90",
		"mnci ci setup", "mnci ci audit", "mnci ci verify", "mnci ci release", "mnci release --dry-run", "PYPI_TOKEN: $(PYPI_TOKEN)", "fetchDepth: 0", "persistCredentials: true", "refs/heads/main"} {
		if !strings.Contains(az, want) {
			t.Errorf("Azure pipeline lacks %q", want)
		}
	}
	// The phase steps are the same in both providers: the lockstep rule.
	phases := func(s string) []string {
		var out []string
		for _, line := range strings.Split(s, "\n") {
			if i := strings.Index(line, "mnci ci "); i >= 0 {
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
	if strings.Contains(gh, "setup-node") || !strings.Contains(gh, "MNCI_VERSION: 1.2.3") || strings.Contains(gh, "env:\n          NODE") {
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
	custom := strings.Replace(string(raw), "      # mnci:slot after-verify\n", "      # mnci:slot after-verify\n      - name: Team step\n        run: make e2e\n", 1)
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
