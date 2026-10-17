package images

import (
	"net"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Login is how release authenticates to a registry.
type Login int

const (
	// LoginPassword logs in with CONTAINER_USERNAME and CONTAINER_PASSWORD.
	LoginPassword Login = iota
	// LoginGitHub logs in to ghcr.io with the job's GITHUB_TOKEN.
	LoginGitHub
	// LoginNone pushes without logging in: a registry on this machine.
	LoginNone
)

// Host is the registry's host (ghcr.io, localhost:5000).
func Host(registry string) string {
	host, _, _ := strings.Cut(registry, "/")
	return host
}

// LoginFor decides how release logs in to a registry.
func LoginFor(registry string) Login {
	host := Host(registry)
	if host == "ghcr.io" {
		return LoginGitHub
	}
	name := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		name = h
	}
	if name == "localhost" || net.ParseIP(name).IsLoopback() {
		return LoginNone
	}
	return LoginPassword
}

// Credentials are the secrets pushing to a target needs, and the GitHub
// token permissions. A registry set with CONTAINER_REGISTRY keeps that
// variable's old, optional credentials.
func Credentials(t Target) ([]protocol.Credential, []string) {
	if t.Legacy {
		return []protocol.Credential{
			{Env: LegacyRegistryEnv, Description: "Registry the apps' images are pushed to; deprecated: set [conventions.docker] registry in smith.toml"},
			{Env: "CONTAINER_USERNAME", Description: "Registry user (omit to use the runner's docker login)"},
			{Env: "CONTAINER_PASSWORD", Description: "Registry password or token"},
		}, nil
	}
	if !t.Push {
		return nil, nil
	}
	switch LoginFor(t.Registry) {
	case LoginGitHub:
		return []protocol.Credential{
			{Env: "GITHUB_TOKEN", Description: "pushes the images to " + t.Registry + " ([conventions.docker] registry): the job's token with packages: write; on Azure Pipelines, a token with write:packages", Required: true},
		}, []string{"packages: write"}
	case LoginNone:
		return nil, nil
	}
	return []protocol.Credential{
		{Env: "CONTAINER_USERNAME", Description: "user that pushes the images to " + Host(t.Registry) + " ([conventions.docker] registry)", Required: true},
		{Env: "CONTAINER_PASSWORD", Description: "password or token that pushes the images to " + Host(t.Registry) + " ([conventions.docker] registry)", Required: true},
	}, nil
}
