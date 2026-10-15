# Plugins

Everything language-specific lives in a plugin: finding projects, their tasks,
scaffolding, writing versions, publishing, CI needs and health checks. mnci's
core knows none of it.

## Built-in plugins

| Plugin | Manifest | Projects | Release |
|---|---|---|---|
| `node` | `package.json` (npm workspaces) | TypeScript; `tsc`, Jest (`ts-jest`), ESLint with `@mnci/eslint-config` (TypeScript rules, Standard style, vertical slices) | `npm publish`, skipping an existing version |
| `go` | `go.mod` (in a root `go.work`) | `go build`, `go test`, `go vet` | a `<dir>/vX.Y.Z` module tag |
| `python` | `pyproject.toml` (src layout) | one root `.venv`, Ruff, pytest, PyPA `build` | `twine upload` of files the index lacks |

Built-in plugins ship inside the `mnci` binary but still run as separate
processes (`mnci plugin serve <name>`), speaking the same protocol an external
plugin does. `mnci plugin list` shows each one's capabilities.

## Writing a plugin

A plugin is any executable that speaks the protocol on stdin/stdout; anything
it wants the user to see goes to stderr. Register it in `mnci.toml`:

```toml
[plugins.external.rust]
command = ["mnci-plugin-rust"]
```

In Go, the SDK (`github.com/MoNecromanCI/mnci/packages/plugin-sdk`) does the
protocol for you:

```go
package main

import (
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/serve"
)

type rust struct{ root string }

func (r *rust) Describe(p protocol.InitializeParams) (protocol.InitializeResult, error) {
	r.root = p.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin: protocol.Peer{Name: "rust", Version: "0.1.0"}, Kind: protocol.KindLanguage,
		Languages: []string{"rust"}, Manifests: []string{"Cargo.toml"},
	}, nil
}

// Implementing Discover is what advertises the discover capability.
func (r *rust) Discover(p protocol.DiscoverParams) (protocol.DiscoverResult, error) { /* ... */ }

func main() { serve.Main(&rust{}) }
```

`host.Connect` drives a plugin in-process over pipes, which is how to test one
(see `packages/plugin-sdk/host/plugin_client_test.go`).

## The protocol (version 1)

JSON-RPC 2.0 over stdio, framed like the Language Server Protocol:

```
Content-Length: 85\r\n
\r\n
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersions":[1],...}}
```

| Method | Params → Result | Capability |
|---|---|---|
| `initialize` | protocol versions, host, workspace root → chosen version, plugin identity, kind (`language`/`convention`), capabilities, languages, manifest names | always |
| `discover` | workspace-relative manifest paths the host found → projects | `discover` |
| `generate` | kind, name, directory, workspace info → files written, notes | `generate` |
| `version/write` | project, version, new dependency versions, dry run → files changed | `version` |
| `publish` | project, version, dry run → published / skipped with a reason | `publish` |
| `ci` | the plugin's projects → toolchains, setup and audit steps, credentials | `ci` |
| `doctor` | the plugin's projects → findings, each with a remedy | `doctor` |
| `shutdown` | → `{}`; the plugin then exits | always |

A **project** carries its name, root, language, kind (`app`, `lib`,
`package`), the identifiers other projects import it by (`exports`), the
identifiers its sources import (`imports` — the host keeps those another
project exports and turns them into graph edges), its manifest version, an
optional release tag prefix, and its tasks.

Rules:

- The host never calls a method the plugin did not advertise. The Go SDK derives
  capabilities from the interfaces a plugin implements, so it cannot lie.
- A breaking change raises the integer protocol version; additions are new
  capabilities or optional fields, which old peers ignore.
- `initialize` must come first. A failure the user must act on returns error
  code `-32000` with a message saying what to do.
- Every call has a timeout; a plugin that does not answer is killed.
- Paths in messages are workspace-relative with forward slashes.

## Convention plugins

A convention describes everything a kind of project does — its tools, rules and
folder structure (for example node + ESLint + Jest + TypeScript + vertical
slices). Today the built-in language plugins carry one convention each; making
conventions separate, composable plugins is epic #11.
