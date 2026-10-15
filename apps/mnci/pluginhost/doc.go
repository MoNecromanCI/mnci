// Package pluginhost starts the workspace's plugins and routes requests to the
// right one. Built-in plugins ship inside the mnci binary but still run
// out-of-process — mnci starts itself as `mnci plugin serve <name>` — so they
// speak exactly the protocol an external plugin does.
package pluginhost
