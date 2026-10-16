<p align="center"><strong>mnci</strong> — a simple monorepo tool</p>

# mnci

Ask for an app, an internal library or a publishable package, in Node, Go or
Python, and it just works: it lints, type-checks, tests, builds and runs at
once, in dependency order, from a cache. CI pipelines are generated from what is
in the repository, and packages are released from conventional commits.

mnci is a single static binary written in Go. Each language is a plugin that
runs out-of-process over a small JSON-RPC protocol, so any language can be added
without touching mnci.

```sh
mnci new acme && cd acme
mnci add app web --lang node        # deployed, never published
mnci add lib core --lang go         # imported internally, never published
mnci add package sdk --lang python  # published on release

mnci build                          # every project, in dependency order, cached
mnci test --affected                # only what changed since main
mnci graph                          # how projects depend on each other
mnci release --dry-run              # what the next release would publish
mnci doctor                         # is the workspace healthy?
```

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/MoNecromanCI/mnci/main/install.sh | sh     # Linux, macOS, Git Bash
irm https://raw.githubusercontent.com/MoNecromanCI/mnci/main/install.ps1 | iex           # Windows PowerShell
```

Or download a zip from the [releases](https://github.com/MoNecromanCI/mnci/releases).
npm, PyPI, Homebrew and Scoop packages are planned (epic #18).

## What it does

| | |
|---|---|
| **Three kinds** | `app` (deployed), `lib` (internal), `package` (published). The kind decides what release does with a project. |
| **No configuration to write** | Projects, tasks and dependencies come from the manifests (`package.json`, `go.mod`, `pyproject.toml`) and from what the sources import. `mnci.toml` names the workspace; everything else has a default. |
| **Cache** | A task whose inputs did not change replays its output and log from a content-addressed cache. |
| **Affected** | `--affected` runs only the projects a change can break, and runs everything when it cannot be sure. |
| **CI generated** | GitHub Actions and Azure Pipelines, derived from the plugins in use, calling `mnci ci <phase>` — the same commands you run locally. |
| **Release** | feat → minor, fix → patch, breaking → major. Packages release in dependency order to npm, PyPI and Go module tags; a failed release resumes when re-run. |
| **Convert** | `mnci convert` turns an Nx workspace, or one made by the old Nx-based mnci, into an mnci workspace, keeping its projects and release tags. It reports anything it cannot map. |
| **Plugins** | Node (TypeScript, ESLint, Jest), Go (modules in a `go.work`), Python (pip, Ruff, pytest). Write your own in any language. |

## Documentation

- [Getting started](docs/getting-started.md)
- [Converting from Nx](docs/converting.md)
- [Configuration (`mnci.toml`)](docs/configuration.md)
- [Tasks, cache and affected](docs/tasks-and-cache.md)
- [CI pipelines](docs/ci.md)
- [Releasing](docs/releasing.md)
- [Stacks: the plugins each project is built with](docs/stacks.md)
- [Conventions (vertical slices)](docs/conventions.md)
- [Plugins and the plugin protocol](docs/plugins.md)
- [Editors: the --json contracts](docs/editors.md)
- [Working on mnci](docs/development.md)
- Background: [decisions](docs/decisions.md) and [research](docs/research/)

## Status

The golden path — a new workspace, every kind of project in Node, Go and
Python, build, test, run, CI and release — works and is tested end to end on
Windows, macOS and Linux. The roadmap is tracked in
[issues](https://github.com/MoNecromanCI/mnci/issues) and on the
[project board](https://github.com/orgs/MoNecromanCI/projects/2).

## License

MIT
