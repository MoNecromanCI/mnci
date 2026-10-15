# Working on mnci

## Layout

This repository was bootstrapped with the current, Nx-based mnci (issue #19)
and is still built by it; the Go tool will take over its own build (#129).

```
apps/mnci/              the CLI: the language-agnostic core, one package per slice
packages/plugin-sdk/    the plugin protocol, framing, runtime, host client and helpers (published)
libs/plugin-node/       the Node language plugin
libs/plugin-go/         the Go language plugin
libs/plugin-python/     the Python language plugin
install.sh, install.ps1 installers used by people and by generated pipelines
docs/                   documentation, decisions and research
```

Go modules are tied together by `go.work`. Code follows vertical slices: a
package per outcome, files named `<name>_<role>.go` (`use_case`, `policy`,
`model`, `contract`, `mapper`, `algorithm`, `client`, `repository`, `store`,
`handler`, `config`, …), and each slice owns its CLI handler.

Core slices in `apps/mnci`:

| Slice | Outcome |
|---|---|
| `workspace` | find the root, load `mnci.toml` |
| `pluginhost` | start plugins, route calls, `mnci plugin` |
| `projectgraph` | projects, edges from imports, overrides with provenance, selection |
| `taskgraph` | expand requested tasks with their dependencies |
| `taskcache` | cache keys, the `Storage` interface and local backend, save/restore/prune |
| `taskrun` | the scheduler, process trees, `mnci run/build/test/lint/typecheck` |
| `affected` | changed projects and their dependents |
| `release` | plan and execute releases, `mnci release` |
| `cipipeline` | pipeline model, renderers, `mnci ci` |
| `scaffold` | `mnci new`, `mnci add` |
| `doctor`, `explain`, `graphview` | inspection commands |
| `session` | open a workspace for a command |
| `gitclient` | every git call |
| `commandline` | assemble the command tree |
| `goldenpath` | the end-to-end test (build tag `e2e`) |

## Checks

```sh
for m in apps/mnci libs/plugin-node libs/plugin-go libs/plugin-python packages/plugin-sdk; do
  (cd $m && go vet ./... && go test ./... && golangci-lint run ./...)
done
node tools/go-slice-check.cjs apps/mnci libs/plugin-node libs/plugin-go libs/plugin-python packages/plugin-sdk

# the golden path, end to end (needs Node 24, Go, Python 3.12 and network)
cd apps/mnci && go test -tags e2e -timeout 40m -v ./goldenpath
```

`npm run build|test|lint` runs the same through the Nx workspace.

## Building

```sh
cd apps/mnci && go build -o ../../dist/dev/mnci .
```

The release build (`nx run mnci:build-all`) stamps `main.version` and builds six
static binaries (`CGO_ENABLED=0`, `-trimpath`).

## Releasing mnci

mnci itself is released by this repository's Nx pipeline: a conventional commit
touching `apps/mnci` on `main` tags `mnci@<version>` and attaches
`mnci_<version>_<os>_<arch>.zip` for six platforms to its GitHub Release, which
is what `install.sh` downloads.
