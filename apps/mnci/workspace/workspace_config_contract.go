package workspace

// FileName is the workspace marker and configuration file.
const FileName = "mnci.toml"

// SchemaVersion is the mnci.toml schema this build reads and writes.
const SchemaVersion = 1

// Config is mnci.toml. Every field is optional except workspace.name.
type Config struct {
	Schema    int              `toml:"schema"`
	Workspace WorkspaceSection `toml:"workspace"`
	CI        CISection        `toml:"ci"`
	Release   ReleaseSection   `toml:"release"`
	Cache     CacheSection     `toml:"cache"`
	Plugins   PluginsSection   `toml:"plugins"`
	// Conventions apply convention plugins, by name, to the projects of the
	// languages they support.
	Conventions map[string]ConventionConfig `toml:"conventions"`
	Tasks       map[string]TaskOverride     `toml:"tasks"`
	Projects    map[string]ProjectOverride  `toml:"projects"`
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
	// GitHubReleases creates a GitHub Release for every tag, with the
	// release's commits as notes and an app's assets attached. Needs
	// GITHUB_TOKEN (or GH_TOKEN). Default: false.
	GitHubReleases bool `toml:"github_releases"`
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

// ConventionConfig chooses the projects a convention applies to.
type ConventionConfig struct {
	// Languages narrows the convention to some of the languages it supports.
	// Default: all of them.
	Languages []string `toml:"languages"`
	// Exclude lists projects the convention leaves alone.
	Exclude []string `toml:"exclude"`
}

// ProjectOverride adjusts one project.
type ProjectOverride struct {
	// TagPrefix replaces the release tag prefix ("<name>@" by default), so a
	// converted project keeps its existing tags.
	TagPrefix string `toml:"tag_prefix"`
	// Release makes an app releasable: it is versioned by its tag and its
	// assets are built and attached to the release.
	Release *ProjectRelease         `toml:"release"`
	Tasks   map[string]TaskOverride `toml:"tasks"`
}

// DefaultPlatforms are the platforms an app's assets are built for.
var DefaultPlatforms = []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "windows/amd64", "windows/arm64"}

// ProjectRelease configures the release of an app.
type ProjectRelease struct {
	// AssetName is the file name pattern of each asset; placeholders are
	// {product}, {version}, {os}, {arch} and {ext}.
	// Default: {product}_{version}_{os}_{arch}.{ext}.
	AssetName string `toml:"asset_name"`
	// Product fills {product}. Default: the project name.
	Product string `toml:"product"`
	// Platforms are os/arch pairs. Default: DefaultPlatforms.
	Platforms []string `toml:"platforms"`
	// OSAlias and ArchAlias rename an os or arch in asset names.
	OSAlias   map[string]string `toml:"os_alias"`
	ArchAlias map[string]string `toml:"arch_alias"`
	// VersionVariable is the symbol the version is stamped into. Default for
	// Go: main.version.
	VersionVariable string `toml:"version_variable"`
}

// TaskOverride replaces individual fields of a plugin-inferred task. Unset
// fields keep the plugin's value.
type TaskOverride struct {
	Command []string `toml:"command"`
	// Cwd is workspace-relative ("." is the workspace root); empty keeps the
	// project root.
	Cwd       string            `toml:"cwd"`
	Inputs    []string          `toml:"inputs"`
	Outputs   []string          `toml:"outputs"`
	DependsOn []string          `toml:"depends_on"`
	Env       map[string]string `toml:"env"`
	Cache     *bool             `toml:"cache"`
}
