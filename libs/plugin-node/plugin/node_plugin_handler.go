package plugin

import (
	"github.com/MoNecromanCI/mnci/libs/plugin-node/discovery"
	"github.com/MoNecromanCI/mnci/libs/plugin-node/pipeline"
	"github.com/MoNecromanCI/mnci/libs/plugin-node/releasing"
	"github.com/MoNecromanCI/mnci/libs/plugin-node/scaffolding"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Node is the Node.js/TypeScript language plugin.
type Node struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Node { return &Node{Version: version} }

// Describe implements serve.Plugin.
func (n *Node) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	n.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: "node", Version: n.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"node", "javascript", "js"},
		Manifests: []string{"package.json"},
	}, nil
}

// Discover implements serve.Discoverer.
func (n *Node) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(n.root, params.Manifests)
}

// Generate implements serve.Generator.
func (n *Node) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(n.root, params)
}

// WriteVersion implements serve.Versioner.
func (n *Node) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(n.root, params)
}

// Publish implements serve.Publisher.
func (n *Node) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(n.root, params)
}

// Published implements serve.RegistryReader.
func (n *Node) Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	return releasing.Published(n.root, params)
}

// CI implements serve.CIContributor.
func (n *Node) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(params, scaffolding.NodeVersion), nil
}

// Doctor implements serve.DoctorContributor.
func (n *Node) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(n.root, params), nil
}

// Remove implements serve.Remover.
func (n *Node) Remove(params protocol.RemoveParams) (protocol.RemoveResult, error) {
	return scaffolding.Remove(n.root, params)
}
