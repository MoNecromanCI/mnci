# plugin-sdk

The Go SDK for writing smith plugins: the protocol's messages, its framing,
the plugin runtime, and the host side used to test a plugin.

```sh
go get github.com/monosmith-app/monosmith/packages/plugin-sdk@latest
```

Released as the Go module tag `packages/plugin-sdk/v<version>`. Before 1.0 a
minor version may change the Go API; the wire protocol is versioned on its own
(protocol version 1), and a breaking change to it raises that integer.

## Packages

| Package | What it is for |
|---|---|
| `protocol` | the JSON-RPC 2.0 messages between smith and a plugin: every method's params and results, project kinds, tasks, stacks, CI contributions |
| `serve` | run a plugin: `serve.Main(plugin)` answers the host on stdin/stdout, deriving the capabilities from the interfaces the plugin implements |
| `transport` | Language Server Protocol framing (`Content-Length`) over a byte stream |
| `host` | the host side: start a plugin, handshake, call with timeouts, shut down; `host.Connect` drives a plugin in-process for tests |
| `conformance` | `conformance.Check`: the checks `smith plugin check` runs, for a plugin's own tests |
| `templating` | render a new project's files from `[[ ]]` templates, never overwriting a file |
| `jsondoc` | edit JSON such as `package.json` in place, keeping key order |
| `projectkind` | a project's kind when its manifest declares none, and the directory each kind lives in |
| `toolrun` | run toolchain commands with their output on stderr, keeping stdout for the protocol |

## A plugin

```go
package main

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

type kiwi struct{}

func (k *kiwi) Describe(p protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{
		Plugin: protocol.Peer{Name: "kiwi", Version: "0.1.0"}, Kind: protocol.KindLanguage,
		Languages: []string{"kiwi"}, Manifests: []string{"kiwi.toml"},
	}, nil
}

func main() { serve.Main(&kiwi{}) }
```

`smith add app my-plugin --stack go-smith-plugin` scaffolds a working plugin
with a conformance test. The protocol, the plugin roles and how a workspace
declares a plugin are in [docs/plugins.md](../../docs/plugins.md);
`apps/smith/goldenpath/testdata/kiwiplugin` is a complete language plugin.
