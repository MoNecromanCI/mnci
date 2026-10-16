<p align="center"><strong>MonoSmith</strong> — a simple monorepo tool</p>

# MonoSmith

MonoSmith runs a monorepo from one command, `smith`. Ask for an app, an internal library or a publishable package, in Node, Go or
Python, and it just works: it lints, type-checks, tests, builds and runs at
once, in dependency order, from a cache. CI pipelines are generated from what is
in the repository, and packages are released from conventional commits.

`smith` is a single static binary written in Go. Each language is a plugin that
runs out-of-process over a small JSON-RPC protocol, so any language can be added
without touching `smith`.

```sh
smith new acme && cd acme
smith add app web --lang node        # deployed, never published
smith add lib core --lang go         # imported internally, never published
smith add package sdk --lang python  # published on release
smith remove web                     # delete a project and its traces (--dry-run first)

smith build                          # every project, in dependency order, cached
smith test --affected                # only what changed since main
smith graph                          # how projects depend on each other (--format tui to browse)
smith release --dry-run              # what the next release would publish
smith doctor                         # is the workspace healthy?
```

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/monosmith-app/monosmith/main/install.sh | sh     # Linux, macOS, Git Bash
irm https://raw.githubusercontent.com/monosmith-app/monosmith/main/install.ps1 | iex           # Windows PowerShell
```

Or with the package manager of your language; each installs the `smith` command:

```sh
npm install --global monosmith
pipx install monosmith         # or: uv tool install monosmith, pip install monosmith
uvx --from monosmith smith --version   # run it without installing
go install github.com/monosmith-app/monosmith/apps/smith@latest
```

Or download a zip from the [releases](https://github.com/monosmith-app/monosmith/releases).
A zip downloaded in a browser is not signed: macOS needs
`xattr -d com.apple.quarantine smith` before the first run, and Windows
SmartScreen asks once. The other ways are not affected.
Homebrew, Scoop, NuGet and pub.dev packages are planned (epic #18).

MonoSmith continues mnci, which was built on Nx; `smith convert` converts its
workspaces.

## What it does

| | |
|---|---|
| **Three kinds** | `app` (deployed), `lib` (internal), `package` (published). The kind decides what release does with a project. |
| **No configuration to write** | Projects, tasks and dependencies come from the manifests (`package.json`, `go.mod`, `pyproject.toml`) and from what the sources import. `smith.toml` names the workspace; everything else has a default. |
| **Cache** | A task whose inputs did not change replays its output and log from a content-addressed cache. |
| **Affected** | `--affected` runs only the projects a change can break, and runs everything when it cannot be sure. |
| **CI generated** | GitHub Actions and Azure Pipelines, derived from the plugins in use, calling `smith ci <phase>` — the same commands you run locally. |
| **Release** | feat → minor, fix → patch, breaking → major. Packages release in dependency order to npm, PyPI and Go module tags; a failed release resumes when re-run. |
| **Convert** | `smith convert` turns an Nx workspace, or one made by the old Nx-based mnci, into a smith workspace, keeping its projects and release tags. It reports anything it cannot map. |
| **Plugins** | Node (TypeScript, ESLint, Jest), Go (modules in a `go.work`), Python (pip, Ruff, pytest). Write your own in any language. |

## Documentation

- [Getting started](docs/getting-started.md)
- [Converting from Nx](docs/converting.md)
- [Configuration (`smith.toml`)](docs/configuration.md)
- [Tasks, cache and affected](docs/tasks-and-cache.md)
- [CI pipelines](docs/ci.md)
- [Releasing](docs/releasing.md)
- [Stacks: the plugins each project is built with](docs/stacks.md)
- [Conventions (vertical slices)](docs/conventions.md)
- [Plugins and the plugin protocol](docs/plugins.md)
- [Editors: the --json contracts](docs/editors.md)
- [Working on MonoSmith](docs/development.md)
- Background: [decisions](docs/decisions.md) and [research](docs/research/)

## Status

The golden path — a new workspace, every kind of project in Node, Go and
Python, build, test, run, CI and release — works and is tested end to end on
Windows, macOS and Linux. The roadmap is tracked in
[issues](https://github.com/monosmith-app/monosmith/issues) and on the
[project board](https://github.com/orgs/MoNecromanCI/projects/2).

## License

MIT
