// Package toolrun runs a plugin's toolchain commands (npm, go, python). A
// plugin's stdout is the protocol channel, so child output is sent to stderr,
// where the host shows it to the user.
package toolrun
