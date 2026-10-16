package plugin

import (
	"github.com/MoNecromanCI/mnci/libs/plugin-docker/images"
	"github.com/MoNecromanCI/mnci/libs/plugin-docker/pipeline"
	"github.com/MoNecromanCI/mnci/libs/plugin-docker/publishing"
	"github.com/MoNecromanCI/mnci/libs/plugin-docker/scaffolding"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Name is the convention's name in mnci.toml ([conventions.docker]).
const Name = "docker"

// languages are the built-in languages: an image ships an app's build,
// whatever built it.
var languages = []string{"node", "go", "python", "rust", "dotnet", "ruby", "php", "cpp", "dart", "jvm"}

// Docker is the docker deploy plugin.
type Docker struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Docker { return &Docker{Version: version} }

// Describe implements serve.Plugin.
func (d *Docker) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	d.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: d.Version},
		Kind:      protocol.KindConvention,
		Roles:     []protocol.Role{protocol.RoleDeploy},
		Languages: languages,
	}, nil
}

// Contribute implements serve.Contributor.
func (d *Docker) Contribute(params protocol.ContributeParams) (protocol.ContributeResult, error) {
	return images.Contribute(d.root, params), nil
}

// Generate implements serve.Generator.
func (d *Docker) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return scaffolding.Generate(d.root, params)
}

// Publish implements serve.Publisher.
func (d *Docker) Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	return publishing.Publish(d.root, params)
}

// CI implements serve.CIContributor.
func (d *Docker) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return pipeline.CI(d.root, params), nil
}

// Doctor implements serve.DoctorContributor.
func (d *Docker) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return pipeline.Doctor(d.root, params), nil
}
