package images

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// LegacyRegistryEnv is the environment variable that named the registry
// before [conventions.docker] registry; still read when the setting is unset.
const LegacyRegistryEnv = "CONTAINER_REGISTRY"

// Target is where release puts an app's image.
type Target struct {
	// Registry is the registry and namespace (ghcr.io/acme), no trailing
	// slash; empty when none is configured.
	Registry string
	// Image is the image's name under the registry.
	Image string
	// Push is false when the image is built but not pushed.
	Push bool
	// Legacy is true when Registry came from CONTAINER_REGISTRY.
	Legacy bool
}

// Reference is the image's reference without a tag (ghcr.io/acme/api).
func (t Target) Reference() string { return t.Registry + "/" + t.Image }

// TargetOf reads the project's docker settings ([conventions.docker]
// registry and push, [projects.<name>.docker] image and push). With no
// registry setting, CONTAINER_REGISTRY from env still names one. Push
// defaults to true when there is a registry.
func TargetOf(p protocol.Project, env func(string) string) Target {
	s := p.Settings["docker"]
	t := Target{Image: Name(p)}
	if registry, _ := s["registry"].(string); registry != "" {
		t.Registry = strings.TrimSuffix(registry, "/")
	} else if legacy := env(LegacyRegistryEnv); legacy != "" {
		t.Registry, t.Legacy = strings.TrimSuffix(legacy, "/"), true
	}
	if image, _ := s["image"].(string); image != "" {
		t.Image = image
	}
	t.Push = t.Registry != ""
	if push, ok := s["push"].(bool); ok {
		t.Push = push && t.Registry != ""
	}
	return t
}

// Configured reports whether smith.toml names the registry (rather than
// CONTAINER_REGISTRY, or nothing).
func Configured(p protocol.Project) bool {
	registry, _ := p.Settings["docker"]["registry"].(string)
	return registry != ""
}

// PushOff reports whether smith.toml turns the push off for the project.
func PushOff(p protocol.Project) bool {
	push, ok := p.Settings["docker"]["push"].(bool)
	return ok && !push
}
