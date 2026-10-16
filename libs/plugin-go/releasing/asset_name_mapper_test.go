package releasing

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAssetName(t *testing.T) {
	spec := protocol.AssetSpec{Name: "{product}_{version}_{os}_{arch}.{ext}", OSAlias: map[string]string{"darwin": "macos"}}
	got, err := AssetName(spec, "smith", "1.2.3", "darwin", "arm64", "zip")
	if err != nil || got != "smith_1.2.3_macos_arm64.zip" {
		t.Fatalf("%s %v", got, err)
	}
	spec.Product = "tool"
	if got, _ := AssetName(spec, "smith", "1.2.3", "linux", "amd64", "zip"); got != "tool_1.2.3_linux_amd64.zip" {
		t.Fatal(got)
	}
	if _, err := AssetName(protocol.AssetSpec{Name: "{name}-{os}"}, "a", "1", "linux", "amd64", "zip"); err == nil || !strings.Contains(err.Error(), "{name}") {
		t.Fatalf("unknown placeholder: %v", err)
	}
}
