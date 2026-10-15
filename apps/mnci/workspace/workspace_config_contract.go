package workspace

// FileName is the workspace marker and configuration file.
const FileName = "mnci.toml"

// SchemaVersion is the mnci.toml schema this build reads and writes.
const SchemaVersion = 1

// Config is mnci.toml. Every field is optional except workspace.name.
type Config struct {
	Schema    int                        `toml:"schema"`
	Workspace WorkspaceSection           `toml:"workspace"`
	CI        CISection                  `toml:"ci"`
	Release   ReleaseSection             `toml:"release"`
	Cache     CacheSection               `toml:"cache"`
	Plugins   PluginsSection             `toml:"plugins"`
	Tasks     map[string]TaskOverride    `toml:"tasks"`
	Projects  map[string]ProjectOverride `toml:"projects"`
}

// WorkspaceSection names the monorepo.
type WorkspaceSection struct {
	Name string `toml:"name"`
	// Scope is the npm scope for packages, e.g. "@acme".
	Scope string `toml:"scope"`
	// ModulePrefix is the Go module path prefix, e.g. github.com/acme/repo.
	ModulePrefix string `toml:"module_prefix"`
}

// CISection chooses which pipelines mnci generates.
type CISection struct {
	// Providers is any of "github", "azure". Default: ["github"].
	Providers []string `toml:"providers"`
}

// ReleaseSection configures `mnci release`.
type ReleaseSection struct {
	// Branch is the only branch a real release runs from. Default: main.
	Branch string `toml:"branch"`
	// Remote receives release tags. Default: origin.
	Remote string `toml:"remote"`
}

// CacheSection configures the local task cache.
type CacheSection struct {
	// Dir is workspace-relative. Default: .mnci/cache.
	Dir string `toml:"dir"`
	// MaxSizeMB triggers pruning of the least recently used entries. Default: 2048.
	MaxSizeMB int `toml:"max_size_mb"`
}

// PluginsSection chooses plugins.
type PluginsSection struct {
	// Enabled lists built-in plugins to load. Default: all of them.
	Enabled []string `toml:"enabled"`
	// External plugins are executables speaking the plugin protocol.
	External map[string]ExternalPlugin `toml:"external"`
}

// ExternalPlugin is a plugin outside the mnci binary.
type ExternalPlugin struct {
	Command []string `toml:"command"`
}

// ProjectOverride adjusts one project.
type ProjectOverride struct {
	Tasks map[string]TaskOverride `toml:"tasks"`
}

// TaskOverride replaces individual fields of a plugin-inferred task. Unset
// fields keep the plugin's value.
type TaskOverride struct {
	Command   []string          `toml:"command"`
	Inputs    []string          `toml:"inputs"`
	Outputs   []string          `toml:"outputs"`
	DependsOn []string          `toml:"depends_on"`
	Env       map[string]string `toml:"env"`
	Cache     *bool             `toml:"cache"`
}
