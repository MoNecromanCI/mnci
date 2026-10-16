# Plugins

Everything language-specific lives in a plugin: finding projects, their tasks,
scaffolding, writing versions, publishing, CI needs and health checks. mnci's
core knows none of it. A plugin plays one or more roles (language, toolchain,
structure, scaffold, deploy), and each project picks the plugins it is built
with through its [stack](stacks.md).

## Built-in plugins

| Plugin | Roles | Does | Release |
|---|---|---|---|
| `node` | language (+ default toolchain, scaffold) | `package.json` projects (npm workspaces); TypeScript, `tsc` builds and type-checks, every package.json script is a task | `npm publish`, skipping an existing version |
| `go` | language (+ default toolchain, scaffold) | `go.mod` projects (in a root `go.work`); `go build`, `go test`, `go vet` | a `<dir>/vX.Y.Z` module tag |
| `python` | language (+ default toolchain, scaffold) | `pyproject.toml` projects (src layout); one root `.venv`, Ruff, pytest, PyPA `build` | `twine upload` of files the index lacks |
| `eslint` | toolchain, provides `lint` | Node projects lint with ESLint and `@mnci/eslint-config` (TypeScript rules, Standard style, vertical slices) | — |
| `jest` | toolchain, provides `test`; publishes `node-jest` | Node projects test with Jest (`ts-jest`) | — |
| `vitest` | toolchain, provides `test`; publishes `node-vitest` | Node projects test with Vitest | — |
| `vertical-slices` | structure | a `slice-check` task for Go and Python projects, run before `lint` | — |

Built-in plugins ship inside the `mnci` binary but still run as separate
processes (`mnci plugin serve <name>`), speaking the same protocol an external
plugin does. `mnci plugin list` shows each one's roles and capabilities, and
`mnci stacks` the stacks they publish.

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
| `initialize` | protocol versions, host, workspace root → chosen version, plugin identity, kind (`language`/`convention`), roles, the tasks it provides, the stacks it publishes, capabilities, languages, manifest names | always |
| `discover` | workspace-relative manifest paths the host found → projects | `discover` |
| `generate` | kind, name, directory, workspace info, stack → files written, notes | `generate` |
| `version/write` | project, version, new dependency versions, dry run → files changed | `version` |
| `publish` | project, version, dry run → published / skipped with a reason | `publish` |
| `ci` | the plugin's projects → toolchains (with their GitHub and Azure setup steps, or an install command), setup and audit steps, credentials | `ci` |
| `doctor` | the plugin's projects → findings, each with a remedy | `doctor` |
| `contribute` | the projects whose stack includes the plugin, or a convention applies to → new tasks, replaced tasks and extra task dependencies, by project | `contribute` |
| `shutdown` | → `{}`; the plugin then exits | always |

A **project** carries its name, root, language, kind (`app`, `lib`,
`package`), the identifiers other projects import it by (`exports`), the
identifiers its sources import (`imports` — the host keeps those another
project exports and turns them into graph edges), its manifest version, an
optional release tag prefix, its tasks, and the stack its manifest records.

Rules:

- The host never calls a method the plugin did not advertise. The Go SDK derives
  capabilities from the interfaces a plugin implements, so it cannot lie.
- A breaking change raises the integer protocol version; additions are new
  capabilities or optional fields, which old peers ignore.
- `initialize` must come first. A failure the user must act on returns error
  code `-32000` with a message saying what to do.
- Every call has a timeout; a plugin that does not answer is killed.
- Paths in messages are workspace-relative with forward slashes.

## Roles

A plugin declares its roles in `initialize` (`roles`; a language plugin that
declares none is language, toolchain and scaffold, any other is structure):

- A **language** plugin implements `discover`, and usually `generate`,
  `version/write` and `publish`. It is what makes a new language work in mnci:
  the graph, the cache, affected and releases need nothing else.
- A **toolchain**, **structure** or **deploy** plugin implements
  `contribute` (`serve.Contributor` in the Go SDK) to add or replace tasks of
  the projects whose stack includes it, `generate` (`serve.Generator`) to add
  its files, scripts and dependencies when such a project is created, and
  `ci` and `doctor` for those projects. A toolchain plugin lists the tasks it
  owns in `provides`, so a stack cannot hold two test runners.
- A **scaffold** plugin implements `generate` and writes a new project in
  place of the language plugin. It records `params.stack` in the manifest the
  way the language plugin does, so discovery reports it back.

A plugin can publish ready-made `stacks` in `initialize`. Any plugin of kind
`convention` can also be applied workspace-wide with `[conventions.<name>]`;
see [conventions](conventions.md). [Stacks](stacks.md) explains how a
project's plugins combine.
