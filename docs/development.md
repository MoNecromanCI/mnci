# Working on mnci

## Layout

This repository was bootstrapped with the old, Nx-based mnci (#19) and
converted with `mnci convert` (#129): it is an mnci workspace, built, tested and
released by the mnci it contains.

```
apps/mnci/              the CLI: the language-agnostic core, one package per slice
packages/plugin-sdk/    the plugin protocol, framing, runtime, host client and helpers (published)
libs/plugin-node/       the Node language plugin
libs/plugin-go/         the Go language plugin
libs/plugin-python/     the Python language plugin
libs/plugin-vertical-slices/    the vertical-slices structure plugin
libs/plugin-node-toolchains/   the typescript, eslint, jest and vitest plugins and their stacks (node-ts, node-jest, node-vitest)
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
| `pluginhost` | find and start plugins (the search path, `plugin.json`), route calls, the stack catalog, `mnci plugin list/check` |
| `projectgraph` | projects, edges from imports, overrides with provenance, selection |
| `taskgraph` | expand requested tasks with their dependencies |
| `taskcache` | cache keys, the `Storage` interface and local backend, save/restore/prune |
| `taskrun` | the scheduler, process trees, `mnci run/build/test/lint/typecheck` |
| `affected` | changed projects and their dependents |
| `release` | plan and execute releases, `mnci release` |
| `releaselineage` | decide whether release tags can be trusted: old names, published but untagged, behind the registry |
| `releasepreview` | post the release preview on the pull request (GitHub, Azure Repos) |
| `cipipeline` | pipeline model, renderers, `mnci ci` |
| `scaffold` | `mnci new`, `mnci add` |
| `nxworkspace`, `convertplan`, `convert` | read an Nx workspace, plan its conversion, apply it (`mnci convert`) |
| `githubrelease` | GitHub Releases and their assets |
| `slicecheck` | `mnci slices check`, the vertical-slices convention's check |
| `doctor`, `explain`, `whymiss`, `stackview`, `graphview` | inspection commands |
| `clicontract` | `mnci commands`, `mnci kinds`, `mnci info`: the --json contracts editors read |
| `session` | open a workspace for a command |
| `gitclient` | every git call |
| `commandline` | assemble the command tree |
| `goldenpath` | the end-to-end tests (build tag `e2e`): the golden path, a language added by an external plugin (`testdata/kiwiplugin`), and converting Nx and old-mnci workspaces |

## Checks

```sh
for m in apps/mnci libs/plugin-node libs/plugin-node-toolchains libs/plugin-go libs/plugin-python libs/plugin-vertical-slices packages/plugin-sdk; do
  (cd $m && go vet ./... && go test ./... && golangci-lint run ./... && go run github.com/MoNecromanCI/mnci/apps/mnci slices check .)
done

# the golden path, end to end (needs Node 24, Go, Python 3.12 and network)
cd apps/mnci && go test -tags e2e -timeout 40m -v ./goldenpath
```

`mnci run lint test build` runs the same through the workspace (lint includes
the vertical-slices convention's `slice-check`), with the cache. Use the mnci you are working on:
`go run ./apps/mnci run lint test build`.

## Building

```sh
cd apps/mnci && go build -o ../../dist/dev/mnci .
```

`mnci release` builds the release binaries: the Go plugin stamps `main.version`
and builds six static binaries (`CGO_ENABLED=0`, `-trimpath`), as configured in
`[projects.mnci.release]` of `mnci.toml`.

## Releasing mnci

mnci releases itself. The generated pipeline builds mnci from the commit (the
`before-setup` slot of `.github/workflows/ci.yml`) and runs `mnci ci release` on
`main`. A conventional commit touching `apps/mnci` (or a lib it bundles) tags
`mnci@<version>` and attaches `mnci_<version>_<os>_<arch>.zip` for six platforms
and `checksums.txt` to its GitHub Release, which is what `install.sh`
downloads. `packages/plugin-sdk` releases as the Go module tag
`packages/plugin-sdk/v<version>`.
