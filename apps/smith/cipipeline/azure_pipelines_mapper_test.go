package cipipeline

import (
	"strings"
	"testing"
)

func TestANativeProjectAddsAJobPerOS(t *testing.T) {
	m := model()
	if strings.Contains(RenderGitHub(m), "native:") || strings.Contains(RenderAzure(m), "jobs:") {
		t.Fatal("no native project: no native job, and Azure keeps its single job")
	}
	m.Native, m.NativeRunners = true, DefaultNativeRunners
	gh, az := RenderGitHub(m), RenderAzure(m)
	for _, want := range []string{
		"\n  native:\n    needs: ci\n    strategy:\n      fail-fast: false\n      matrix:\n        os: [ubuntu-latest, windows-latest, macos-latest]\n    runs-on: ${{ matrix.os }}\n",
		"      - uses: actions/setup-node@v4\n",
		"      - name: Install mnci\n        shell: bash\n",
		"      - name: Native\n        run: mnci ci native\n        env:\n          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}\n",
		"          name: native-${{ matrix.os }}\n",
	} {
		if !strings.Contains(gh, want) {
			t.Errorf("GitHub workflow lacks %q", want)
		}
	}
	for _, want := range []string{
		"\njobs:\n  - job: ci\n    pool:\n      vmImage: ubuntu-latest\n    steps:\n      - checkout: self\n",
		"      # mnci:slot before-setup\n",
		"      - bash: mnci ci verify\n",
		"  - job: native\n    dependsOn: ci\n    strategy:\n      matrix:\n        ubuntu_latest:\n          image: ubuntu-latest\n        windows_latest:\n          image: windows-latest\n",
		"    pool:\n      vmImage: $(image)\n",
		"      - task: UseNode@1\n",
		"      - bash: mnci ci native\n",
		"        artifact: native-$(image)\n",
	} {
		if !strings.Contains(az, want) {
			t.Errorf("Azure pipeline lacks %q", want)
		}
	}
	if strings.Contains(az, "\nsteps:\n") || strings.Contains(az, "\npool:\n") {
		t.Error("with jobs, Azure has no top-level steps or pool")
	}
}

// A team step kept in a slot follows the slot when the Azure pipeline nests
// its steps under jobs.
func TestSlotsAreReindentedWhenTheyMove(t *testing.T) {
	m := model()
	old := strings.Replace(RenderAzure(m), "  # mnci:slot before-setup\n", "  # mnci:slot before-setup\n  - bash: make tools\n    displayName: Team tools\n", 1)
	m.Native, m.NativeRunners = true, DefaultNativeRunners
	kept, lost := KeepSlots(old, RenderAzure(m))
	if len(lost) != 0 || !strings.Contains(kept, "      # mnci:slot before-setup\n      - bash: make tools\n        displayName: Team tools\n      # mnci:slot-end before-setup") {
		t.Fatalf("lost %v:\n%s", lost, kept)
	}
}
