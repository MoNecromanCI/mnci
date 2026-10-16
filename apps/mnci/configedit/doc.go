// Package configedit reads and changes single settings of mnci.toml: `mnci
// config list|get|set|unset`, and the TUI's settings view. Values are checked
// against mnci.toml's JSON Schema, and the file is edited as text, so
// comments and every other line survive.
package configedit
