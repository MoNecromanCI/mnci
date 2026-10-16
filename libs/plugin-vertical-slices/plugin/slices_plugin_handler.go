package plugin

import (
	"github.com/monosmith-app/monosmith/libs/plugin-vertical-slices/languages"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Name is the convention's name in smith.toml ([conventions.vertical-slices]).
const Name = "vertical-slices"

// VerticalSlices is the vertical-slices convention plugin.
type VerticalSlices struct {
	Version string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *VerticalSlices { return &VerticalSlices{Version: version} }

// Describe implements serve.Plugin.
func (v *VerticalSlices) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: v.Version},
		Kind:      protocol.KindConvention,
		Languages: languages.Names(),
	}, nil
}

// Contribute implements serve.Contributor: a slice-check task for every
// project, run before lint.
func (v *VerticalSlices) Contribute(params protocol.ContributeParams) (protocol.ContributeResult, error) {
	res := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, p := range params.Projects {
		rules, ok := languages.For(p.Language)
		if !ok {
			continue
		}
		c := protocol.Contribution{Tasks: map[string]protocol.Task{TaskName: sliceCheck(rules)}}
		if _, ok := p.Tasks["lint"]; ok {
			c.DependsOn = map[string][]string{"lint": {TaskName}}
		}
		res.Projects[p.Name] = c
	}
	return res, nil
}
