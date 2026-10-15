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
)
