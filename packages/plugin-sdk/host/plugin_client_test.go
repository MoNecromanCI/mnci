package host_test

import (
	"errors"
	"io"
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/host"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/serve"
)

type fakePlugin struct{}

func (fakePlugin) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{Plugin: protocol.Peer{Name: "fake", Version: "1.0.0"}, Kind: protocol.KindLanguage, Manifests: []string{"fake.json"}}, nil
}

func (fakePlugin) Discover(p protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	var out protocol.DiscoverResult
	for _, m := range p.Manifests {
		if m == "bad/fake.json" {
			return out, errors.New("bad manifest")
		}
		out.Projects = append(out.Projects, protocol.Project{Name: m, Root: m, Kind: protocol.ProjectLib})
	}
	return out, nil
}

func connect(t *testing.T) *host.Client {
	t.Helper()
	hostR, pluginW := io.Pipe()
	pluginR, hostW := io.Pipe()
	go func() { _ = serve.Serve(fakePlugin{}, pluginR, pluginW); _ = pluginW.Close() }()
	c, err := host.Connect(hostR, hostW, host.Options{WorkspaceRoot: "/ws"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestHandshakeAndCapabilities(t *testing.T) {
	c := connect(t)
	if c.Info.Plugin.Name != "fake" || c.Info.ProtocolVersion != protocol.Version {
		t.Fatalf("unexpected info %+v", c.Info)
	}
	if !c.Info.Capabilities.Discover || c.Info.Capabilities.Generate {
		t.Fatalf("capabilities must mirror implemented interfaces: %+v", c.Info.Capabilities)
	}
	if !c.Supports(protocol.MethodDiscover) || c.Supports(protocol.MethodPublish) {
		t.Fatal("Supports disagrees with capabilities")
	}
}

func TestCallsAndErrors(t *testing.T) {
	c := connect(t)
	var res protocol.DiscoverResult
	if err := c.Call(protocol.MethodDiscover, protocol.DiscoverParams{Manifests: []string{"a/fake.json"}}, &res, 0); err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 1 || res.Projects[0].Name != "a/fake.json" {
		t.Fatalf("unexpected result %+v", res)
	}
	err := c.Call(protocol.MethodDiscover, protocol.DiscoverParams{Manifests: []string{"bad/fake.json"}}, &res, 0)
	var rpcErr *protocol.ResponseError
	if !errors.As(err, &rpcErr) || rpcErr.Code != protocol.CodePluginFailure {
		t.Fatalf("want plugin failure, got %v", err)
	}
	err = c.Call(protocol.MethodPublish, struct{}{}, nil, 0)
	if !errors.As(err, &rpcErr) || rpcErr.Code != protocol.CodeMethodNotFound {
		t.Fatalf("want method not found, got %v", err)
	}
}

func TestRejectsUnknownProtocolVersion(t *testing.T) {
	hostR, pluginW := io.Pipe()
	pluginR, hostW := io.Pipe()
	go func() { _ = serve.Serve(fakePlugin{}, pluginR, pluginW); _ = pluginW.Close() }()
	c, err := host.Connect(hostR, hostW, host.Options{})
	if err != nil {
		t.Fatal(err)
	}
	var res protocol.InitializeResult
	err = c.Call(protocol.MethodInitialize, protocol.InitializeParams{ProtocolVersions: []int{99}}, &res, 0)
	if err == nil {
		t.Fatal("expected a version negotiation error")
	}
	_ = c.Close()
}
