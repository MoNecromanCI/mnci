package plugin

import (
	"github.com/MoNecromanCI/mnci/libs/plugin-python/dependencies"
	"github.com/MoNecromanCI/mnci/libs/plugin-python/discovery"
	"github.com/MoNecromanCI/mnci/libs/plugin-python/pipeline"
	"github.com/MoNecromanCI/mnci/libs/plugin-python/releasing"
	"github.com/MoNecromanCI/mnci/libs/plugin-python/scaffolding"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Python is the pip-native Python language plugin.
type Python struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Python { return &Python{Version: version} }

// Describe implements serve.Plugin.
func (p *Python) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	p.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: "python", Version: p.Version},
		Kind:      protocol.KindLanguage,
		Languages: []string{"python", "py"},
		Manifests: []string{"pyproject.toml"},
	}, nil
}

// Discover implements serve.Discoverer.
func (p *Python) Discover(params protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	return discovery.Discover(p.root, params.Manifests)
}

// Generate implements serve.Generator.
func (p *Python) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(p.root, params)
}

// WriteVersion implements serve.Versioner.
func (p *Python) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return releasing.WriteVersion(p.root, params)
}

// Publish implements serve.Publisher.
func (p *Python) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return releasing.Publish(p.root, params)
}

// Published implements serve.RegistryReader.
func (p *Python) Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	return releasing.Published(params)
}

// CI implements serve.CIContributor.
func (p *Python) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(params), nil
}

// Doctor implements serve.DoctorContributor.
func (p *Python) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(p.root, params), nil
}

// Remove implements serve.Remover.
func (p *Python) Remove(params protocol.RemoveParams) (protocol.RemoveResult, error) {
	return scaffolding.Remove(p.root, params)
}

// Install implements serve.DependencyManager.
func (p *Python) Install(params protocol.InstallParams) (protocol.DependencyResult, error) {
	return dependencies.Install(p.root, params)
}

// Uninstall implements serve.DependencyManager.
func (p *Python) Uninstall(params protocol.UninstallParams) (protocol.DependencyResult, error) {
	return dependencies.Uninstall(p.root, params)
}
