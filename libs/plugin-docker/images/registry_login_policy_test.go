package images

import (
	"strings"
	"testing"
)

func TestLoginFor(t *testing.T) {
	for registry, want := range map[string]Login{
		"ghcr.io/acme":              LoginGitHub,
		"registry.example.com/team": LoginPassword,
		"acme.azurecr.io":           LoginPassword,
		"localhost:5000/e2e":        LoginNone,
		"127.0.0.1:32768/smith-e2e": LoginNone,
		"[::1]:5000/e2e":            LoginNone,
	} {
		if got := LoginFor(registry); got != want {
			t.Errorf("%s: %v, want %v", registry, got, want)
		}
	}
}

func TestCredentialsFollowTheRegistry(t *testing.T) {
	creds, perms := Credentials(Target{Registry: "ghcr.io/acme", Push: true})
	if len(creds) != 1 || creds[0].Env != "GITHUB_TOKEN" || !creds[0].Required || len(perms) != 1 || perms[0] != "packages: write" {
		t.Errorf("ghcr.io: %+v %v", creds, perms)
	}
	creds, perms = Credentials(Target{Registry: "registry.example.com/team", Push: true})
	if len(creds) != 2 || creds[0].Env != "CONTAINER_USERNAME" || !creds[1].Required || perms != nil ||
		!strings.Contains(creds[1].Description, "[conventions.docker] registry") {
		t.Errorf("another registry: %+v %v", creds, perms)
	}
	if creds, _ := Credentials(Target{Registry: "localhost:5000/e2e", Push: true}); creds != nil {
		t.Errorf("a local registry needs no login: %+v", creds)
	}
	if creds, _ := Credentials(Target{Registry: "ghcr.io/acme"}); creds != nil {
		t.Errorf("no push, no credentials: %+v", creds)
	}
	creds, _ = Credentials(Target{Registry: "ghcr.io/acme", Push: true, Legacy: true})
	if len(creds) != 3 || creds[0].Env != LegacyRegistryEnv || creds[0].Required {
		t.Errorf("CONTAINER_REGISTRY keeps its optional credentials: %+v", creds)
	}
}
