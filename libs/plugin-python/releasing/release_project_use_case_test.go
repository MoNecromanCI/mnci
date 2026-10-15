package releasing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestPublishUploadsOnlyThisVersion(t *testing.T) {
	ws := t.TempDir()
	dist := filepath.Join(ws, "packages", "sdk", "dist")
	_ = os.MkdirAll(dist, 0o755)
	for _, f := range []string{"acme_sdk-1.0.0-py3-none-any.whl", "acme_sdk-1.0.0.tar.gz", "acme_sdk-0.9.0.tar.gz"} {
		_ = os.WriteFile(filepath.Join(dist, f), nil, 0o644)
	}
	var argv []string
	Twine = func(_ string, a []string, _ []string) (string, error) {
		argv = a
		return "Uploading acme_sdk-1.0.0\n", nil
	}
	res, err := Publish(ws, protocol.PublishParams{Project: protocol.Project{Root: "packages/sdk"}, Version: "1.0.0"})
	if err != nil || !res.Published {
		t.Fatalf("res %+v err %v", res, err)
	}
	joined := strings.Join(argv, " ")
	if strings.Contains(joined, "0.9.0") || !strings.Contains(joined, "--skip-existing") || strings.Count(joined, "1.0.0") != 2 {
		t.Fatalf("argv %v", argv)
	}

	Twine = func(string, []string, []string) (string, error) {
		return "Skipping acme_sdk-1.0.0 because it appears to already exist\n", nil
	}
	res, _ = Publish(ws, protocol.PublishParams{Project: protocol.Project{Root: "packages/sdk"}, Version: "1.0.0"})
	if res.Published || res.Skipped != "already published" {
		t.Fatalf("res %+v", res)
	}
	if _, err := Publish(ws, protocol.PublishParams{Project: protocol.Project{Root: "packages/sdk"}, Version: "2.0.0"}); err == nil {
		t.Fatal("publishing without a built distribution must fail")
	}
}
