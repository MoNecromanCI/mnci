// Package host is the host side of the plugin protocol: it starts a plugin
// process, performs the initialize handshake, calls methods with a timeout and
// shuts the plugin down. It is also the test kit for plugin authors: Connect
// drives a plugin served in-process over pipes.
package host
