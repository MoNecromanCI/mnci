package discovery

import (
	"sort"

	"github.com/pelletier/go-toml/v2"
)

// Crate is what mnci reads from a Cargo.toml.
type Crate struct {
	// Name is the package name; empty for a virtual (workspace-only)
	// manifest.
	Name    string
	Version string
	// Kind and Stack are [package.metadata.mnci] kind and stack.
	Kind, Stack string
	// PathDependencies are the crates it depends on by path (the workspace's
	// own), by package name.
	PathDependencies []string
	// Members are [workspace] members, for a workspace root.
	Members []string
}

type manifest struct {
	Package *struct {
		Name     string `toml:"name"`
		Version  any    `toml:"version"`
		Metadata struct {
			Mnci struct {
				Kind  string `toml:"kind"`
				Stack string `toml:"stack"`
			} `toml:"mnci"`
		} `toml:"metadata"`
	} `toml:"package"`
	Workspace *struct {
		Members []string `toml:"members"`
	} `toml:"workspace"`
	Dependencies      map[string]any `toml:"dependencies"`
	DevDependencies   map[string]any `toml:"dev-dependencies"`
	BuildDependencies map[string]any `toml:"build-dependencies"`
}

// ParseCargo reads a Cargo.toml.
func ParseCargo(src []byte) (Crate, error) {
	var m manifest
	if err := toml.Unmarshal(src, &m); err != nil {
		return Crate{}, err
	}
	var c Crate
	if m.Package != nil {
		c.Name = m.Package.Name
		if v, ok := m.Package.Version.(string); ok {
			c.Version = v
		}
		c.Kind, c.Stack = m.Package.Metadata.Mnci.Kind, m.Package.Metadata.Mnci.Stack
	}
	if m.Workspace != nil {
		c.Members = m.Workspace.Members
	}
	seen := map[string]bool{}
	for _, deps := range []map[string]any{m.Dependencies, m.DevDependencies, m.BuildDependencies} {
		for key, raw := range deps {
			table, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if _, local := table["path"]; !local {
				continue
			}
			name := key
			if pkg, ok := table["package"].(string); ok && pkg != "" {
				name = pkg
			}
			if !seen[name] {
				seen[name] = true
				c.PathDependencies = append(c.PathDependencies, name)
			}
		}
	}
	sort.Strings(c.PathDependencies)
	return c, nil
}
