package plugin

import (
	"github.com/MoNecromanCI/mnci/libs/plugin-go/discovery"
	"github.com/MoNecromanCI/mnci/libs/plugin-go/pipeline"
	"github.com/MoNecromanCI/mnci/libs/plugin-go/releasing"
	"github.com/MoNecromanCI/mnci/libs/plugin-go/scaffolding"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Go is the Go language plugin.
type Go struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Go { return &Go{Version: version} }

// Describe implements serve.Plugin.
func (g *Go) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	g.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: "go", Version: g.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"go", "golang"},
		Manifests: []string{"go.mod"},
	}, nil
}

// Discover implements serve.Discoverer.
func (g *Go) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(g.root, params.Manifests)
}

// Generate implements serve.Generator.
func (g *Go) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(g.root, params)
}

// Publish implements serve.Publisher.
func (g *Go) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(g.root, params)
}

// CI implements serve.CIContributor.
func (g *Go) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(g.root, params), nil
}

// Doctor implements serve.DoctorContributor.
func (g *Go) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(g.root, params), nil
}

// Remove implements serve.Remover.
func (g *Go) Remove(params protocol.RemoveParams) (protocol.RemoveResult, error) {
	return pipeline.Remove(g.root, params)
}
