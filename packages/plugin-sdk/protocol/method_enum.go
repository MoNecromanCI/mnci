package protocol

// Method names of protocol version 1. A host never calls a method the plugin
// did not advertise in its Capabilities.
const (
	MethodInitialize = "initialize"
	MethodShutdown   = "shutdown"
	MethodDiscover   = "discover"
	MethodGenerate   = "generate"
	MethodVersion    = "version/write"
	MethodPublish    = "publish"
	MethodCI         = "ci"
	MethodDoctor     = "doctor"
	// MethodContribute asks a convention plugin for its additions.
	MethodContribute = "contribute"
	// MethodRemove lets a plugin clean up after a project is removed.
	MethodRemove = "remove"
	// MethodPublished asks for the newest version a registry has of a
	// project, so release can check its tags against it.
	MethodPublished = "published"
)
