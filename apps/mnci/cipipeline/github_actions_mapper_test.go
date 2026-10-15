package cipipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/apps/mnci/workspace"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func model() Model {
	var c Contributions
	c.Add(protocol.CIResult{Toolchains: []protocol.Toolchain{{Name: "python", Version: "3.12"}}, Credentials: []protocol.Credential{{Env: "PYPI_TOKEN", Required: true}}})
	c.Add(protocol.CIResult{Toolchains: []protocol.Toolchain{{Name: "node", Version: "24"}, {Name: "python", Version: "9"}}, Credentials: []protocol.Credential{{Env: "NODE_AUTH_TOKEN"}}})
	c.Add(protocol.CIResult{Toolchains: []protocol.Toolchain{{Name: "go", Version: "1.26"}}})
	return NewModel(c, "dev", "main")
}

func TestRenderersCarryTheSameSteps(t *testing.T) {
	m := model()
	gh, az := RenderGitHub(m), RenderAzure(m)
	for _, want := range []string{"node-version: '24'", "python-version: '3.12'", "go-version: '1.26'", "mnci ci setup", "mnci ci audit", "mnci ci verify", "mnci ci release", "mnci release --dry-run", "NODE_AUTH_TOKEN: ${{ secrets.NODE_AUTH_TOKEN }}", "fetch-depth: 0", "MNCI_VERSION: latest", "github.ref_name == 'main'"} {
		if !strings.Contains(gh, want) {
			t.Errorf("GitHub workflow lacks %q", want)
		}
	}
	for _, want := range []string{"version: '24.x'", "versionSpec: '3.12'", "version: '1.26.0'", "mnci ci setup", "mnci ci audit", "mnci ci verify", "mnci ci release", "mnci release --dry-run", "PYPI_TOKEN: $(PYPI_TOKEN)", "fetchDepth: 0", "persistCredentials: true", "refs/heads/main"} {
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
