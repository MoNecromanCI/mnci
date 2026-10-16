// Package mcpserver is `mnci mcp`: the workspace served to an AI assistant
// over the Model Context Protocol (JSON-RPC on stdio). Each tool is an mnci
// command line: the server runs mnci itself with it, so a tool does exactly
// what the command does, and a client can show the user that command.
package mcpserver
