package tooling

import "github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"

// Convention serves one Node convention over the plugin protocol: when a Node
// project is generated, it applies the Setup the convention describes.
type Convention struct {
	Name    string
	Version string
	// Setup returns what to add to the project params names.
	Setup func(workspaceRoot string, params protocol.GenerateParams) Setup
	root  string
}

// Describe implements serve.Plugin.
func (c *Convention) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	c.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: c.Name, Version: c.Version},
		Kind:      protocol.KindConvention,
		Languages: []string{"node"},
	}, nil
}

// Generate implements serve.Generator: it runs after the node plugin has
// generated the project.
func (c *Convention) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return Apply(c.root, params, c.Setup(c.root, params))
}
