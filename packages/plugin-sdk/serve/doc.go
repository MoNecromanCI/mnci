// Package serve runs a plugin: it answers the host's requests on stdin/stdout
// by calling the optional interfaces the plugin implements. Capabilities are
// derived from those interfaces, so a plugin cannot advertise a method it
// lacks. Anything a plugin wants the user to see goes to stderr.
package serve
