package images

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestTargetOf(t *testing.T) {
	none := func(string) string { return "" }
	legacy := func(k string) string {
		if k == LegacyRegistryEnv {
			return "registry.example.com/team/"
		}
		return ""
	}
	api := protocol.Project{Name: "API", Kind: protocol.ProjectApp}
	if got := TargetOf(api, none); got.Push || got.Registry != "" || got.Image != "api" {
		t.Errorf("no registry builds without pushing: %+v", got)
	}
	if got := TargetOf(api, legacy); !got.Push || !got.Legacy || got.Reference() != "registry.example.com/team/api" {
		t.Errorf("CONTAINER_REGISTRY still names one: %+v", got)
	}
	api.Settings = map[string]map[string]any{"docker": {"registry": "ghcr.io/acme/", "image": "acme-api"}}
	if got := TargetOf(api, legacy); !got.Push || got.Legacy || got.Reference() != "ghcr.io/acme/acme-api" {
		t.Errorf("the setting wins over the env: %+v", got)
	}
	api.Settings["docker"]["push"] = false
	if got := TargetOf(api, none); got.Push || !PushOff(api) || !Configured(api) {
		t.Errorf("push = false: %+v", got)
	}
}
