// Package transport frames JSON-RPC messages over a byte stream the way the
// Language Server Protocol does: a Content-Length header, a blank line, then
// exactly that many bytes of JSON. It works identically over pipes on every OS.
package transport
