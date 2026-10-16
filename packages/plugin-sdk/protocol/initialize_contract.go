package protocol

// InitializeParams is sent once, first, by the host.
type InitializeParams struct {
	// ProtocolVersions lists every major version the host can speak.
	ProtocolVersions []int `json:"protocolVersions"`
	Host             Peer  `json:"host"`
	// WorkspaceRoot is an absolute path using the host OS separators.
	WorkspaceRoot string `json:"workspaceRoot"`
}

// InitializeResult tells the host who the plugin is and what it can do.
type InitializeResult struct {
	// ProtocolVersion is the version the plugin chose from the host's list.
	ProtocolVersion int          `json:"protocolVersion"`
	Plugin          Peer         `json:"plugin"`
	Kind            PluginKind   `json:"kind"`
	Capabilities    Capabilities `json:"capabilities"`
	// Languages are the values a user passes to --lang to reach this plugin.
	Languages []string `json:"languages,omitempty"`
	// Manifests are file base names (package.json, go.mod) whose presence in a
	// directory may make it a project of this plugin.
	Manifests []string `json:"manifests,omitempty"`
	// Roles are what the plugin can be in a stack. Empty means the default
	// of its kind (see DefaultRoles).
	Roles []Role `json:"roles,omitempty"`
	// Provides are the task names a toolchain plugin owns ("test" for a test
	// runner). Two plugins of one stack may not provide the same task.
	Provides []string `json:"provides,omitempty"`
	// Stacks are ready-made stacks the plugin publishes.
	Stacks []Stack `json:"stacks,omitempty"`
}

// EffectiveRoles are the plugin's roles, or the default of its kind.
func (r InitializeResult) EffectiveRoles() []Role {
	if len(r.Roles) > 0 {
		return r.Roles
	}
	return DefaultRoles(r.Kind)
}

// Peer names one side of the connection.
type Peer struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// PluginKind separates plugins that understand a language from plugins that
// add to projects a language plugin discovered (toolchains, structure rules,
// scaffolds, deploy targets), which are all of kind convention.
type PluginKind string

// Plugin kinds.
const (
	KindLanguage   PluginKind = "language"
	KindConvention PluginKind = "convention"
)

// Capabilities are the optional methods a plugin implements.
type Capabilities struct {
	Discover bool `json:"discover,omitempty"`
	Generate bool `json:"generate,omitempty"`
	Version  bool `json:"version,omitempty"`
	Publish  bool `json:"publish,omitempty"`
	CI       bool `json:"ci,omitempty"`
	Doctor   bool `json:"doctor,omitempty"`
	// Contribute is a convention plugin's: it adds tasks to projects other
	// plugins discovered.
	Contribute bool `json:"contribute,omitempty"`
}
