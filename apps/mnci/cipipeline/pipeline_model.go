package cipipeline

import (
	"sort"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// InstallScriptURL serves the mnci installer used by generated pipelines.
const InstallScriptURL = "https://raw.githubusercontent.com/MoNecromanCI/mnci/main/install.sh"

// Model is a pipeline independent of any CI provider.
type Model struct {
	// MnciVersion is installed by the pipeline; "latest" when mnci is a dev build.
	MnciVersion   string
	ReleaseBranch string
	Toolchains    []protocol.Toolchain
	Credentials   []protocol.Credential
	// Pack is true when a project has a package task: pushes to the release
	// branch then run mnci ci pack and keep dist/drop as the "drop" artifact.
	Pack bool
}

// Contributions are every plugin's CI needs for this workspace.
type Contributions struct {
	Toolchains  []protocol.Toolchain
	Setup       []protocol.Step
	Audit       []protocol.Step
	Credentials []protocol.Credential
	// Pack is true when a project has a package task.
	Pack bool
}

// Add merges one plugin's contribution, keeping the first of duplicates.
func (c *Contributions) Add(r protocol.CIResult) {
	for _, t := range r.Toolchains {
		if !hasToolchain(c.Toolchains, t.Name) {
			c.Toolchains = append(c.Toolchains, t)
		}
	}
	c.Setup = append(c.Setup, r.Setup...)
	c.Audit = append(c.Audit, r.Audit...)
	for _, cred := range r.Credentials {
		dup := false
		for _, have := range c.Credentials {
			dup = dup || have.Env == cred.Env
		}
		if !dup {
			c.Credentials = append(c.Credentials, cred)
		}
	}
}

// NewModel builds the model, sorted so the YAML is deterministic.
func NewModel(c Contributions, mnciVersion, releaseBranch string) Model {
	if mnciVersion == "" || mnciVersion == "dev" {
		mnciVersion = "latest"
	}
	m := Model{MnciVersion: mnciVersion, ReleaseBranch: releaseBranch, Pack: c.Pack}
	m.Toolchains = append(m.Toolchains, c.Toolchains...)
	sort.Slice(m.Toolchains, func(i, j int) bool { return m.Toolchains[i].Name < m.Toolchains[j].Name })
	m.Credentials = append(m.Credentials, c.Credentials...)
	sort.Slice(m.Credentials, func(i, j int) bool { return m.Credentials[i].Env < m.Credentials[j].Env })
	return m
}

func hasToolchain(list []protocol.Toolchain, name string) bool {
	for _, t := range list {
		if t.Name == name {
			return true
		}
	}
	return false
}
