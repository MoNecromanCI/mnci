package pluginkit

import (
	"fmt"

	"github.com/MoNecromanCI/mnci/libs/plugin-go/scaffolding"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Name is the scaffold's name, and its stack's.
const Name = "go-mnci-plugin"

// Scaffold generates mnci plugins written in Go.
type Scaffold struct {
	Version string
	root    string
}

// New returns the scaffold; version is reported in the handshake.
func New(version string) *Scaffold { return &Scaffold{Version: version} }

// Describe implements serve.Plugin.
func (s *Scaffold) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	s.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: s.Version},
		Kind:      protocol.KindConvention,
		Roles:     []protocol.Role{protocol.RoleScaffold},
		Languages: []string{"go"},
		Stacks: []protocol.Stack{{
			Name: Name, Description: "an mnci plugin written in Go, on the plugin SDK, with a conformance test",
			Language: "go", Scaffold: Name,
		}},
	}, nil
}

// Generate implements serve.Generator. A plugin is an executable, so only an
// app can be one.
func (s *Scaffold) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if params.Kind != protocol.ProjectApp {
		return protocol.GenerateResult{}, fmt.Errorf("an mnci plugin is an executable: add it as an app (mnci add app %s --stack %s)", params.Name, Name)
	}
	res, err := scaffolding.GenerateFrom(s.root, params, templates, []string{SDK})
	if err == nil {
		res.Notes = append(res.Notes, "Declare it in mnci.toml to use it: see "+params.Directory+"/README.md.")
	}
	return res, err
}
