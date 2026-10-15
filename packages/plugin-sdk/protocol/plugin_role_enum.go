package protocol

// Role is a part a plugin plays in a stack.
type Role string

// Plugin roles. A project has exactly one language and at most one scaffold;
// it may have any number of the others.
const (
	// RoleLanguage finds projects from manifests, turns imports into graph
	// edges, writes versions and publishes.
	RoleLanguage Role = "language"
	// RoleToolchain decides how a project builds, tests and lints.
	RoleToolchain Role = "toolchain"
	// RoleStructure checks layout and code rules.
	RoleStructure Role = "structure"
	// RoleScaffold writes a new project's starter files.
	RoleScaffold Role = "scaffold"
	// RoleDeploy adds pipeline steps and deploy targets.
	RoleDeploy Role = "deploy"
)

// DefaultRoles are the roles of a plugin that declares none: a language
// plugin brings its own default toolchain and scaffold; a convention plugin
// is a structure rule.
func DefaultRoles(kind PluginKind) []Role {
	if kind == KindLanguage {
		return []Role{RoleLanguage, RoleToolchain, RoleScaffold}
	}
	return []Role{RoleStructure}
}
