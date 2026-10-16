# Getting started

## Prerequisites

- **git**, always.
- The toolchain of each language you use: **Node 24** (with npm), **Go 1.26**,
  **Python 3.12+**. mnci tells you what is missing.

## Create a workspace

```sh
mnci new acme
cd acme
```

This writes `mnci.toml`, `.gitignore`, `.gitattributes`, a README and a CI
pipeline, initialises git on `main` and makes the first commit. In a clone that
already exists (with a LICENSE, say), run `mnci new acme --dir .`: files mnci
does not own are kept.

Options: `--scope @acme` (npm scope), `--module-prefix github.com/acme/acme`
(Go module prefix; by default it is derived from the `origin` remote) and
`--ci github,azure`.

## Add projects

```sh
mnci add app web --lang node
mnci add lib ui --lang node
mnci add package sdk --lang node
```

| Kind | Directory | Release |
|---|---|---|
| `app` | `apps/<name>` | never published; deployed by your pipeline |
| `lib` | `libs/<name>` | never published; imported by other projects |
| `package` | `packages/<name>` | published when it has releasable commits |

`--lang` picks the plugin (`node`, `go`, `python`). Once a workspace has only
one language, it is the default.

Each project starts with one worked example slice (`greeting`) laid out as
vertical slices, and passes lint, type-check, test and build immediately.

To take a project out again:

```sh
mnci remove web --dry-run    # what would change
mnci remove web              # asks first; --yes in scripts
```

It deletes the project's directory, lets its plugin clean what it owns
elsewhere (the `go.work` entry, the npm lockfile, the Python venv), drops the
project's tables from `mnci.toml` and regenerates the pipelines. A project
another one imports is refused unless `--force`; its release tags are kept.

## Use one project from another

Just import it — no manifest to edit:

```ts
// apps/web/src/main.ts
import { greet } from '@acme/ui'
```

```go
// apps/api/main.go
import "example.com/acme/libs/core/greeting"
```

```python
# apps/worker/src/acme_worker/__main__.py
from acme_pyutil.greeting import greet
```

npm workspaces, `go.work` and the workspace's Python venv (editable installs)
make every project importable by every other; mnci reads the imports to learn
the dependency graph:

```sh
mnci graph                         # a tree in the terminal
mnci graph --format tui            # browse it: what a project uses, what uses it
mnci graph --focus web             # web, everything it uses and everything that uses it
mnci graph --affected              # what changes since main affect, the changed ones marked
mnci graph --format html -o graph.html   # a page to share; also json and dot
```

Every format shows the same projects; `--focus` and `--affected` narrow all
of them the same way.

## Add dependencies

Each project owns its dependencies, so adding one names the project:

```sh
mnci install -w web zod                    # npm install zod --workspace apps/web
mnci install -w worker -D pytest-cov       # pyproject.toml's dev extra, into .venv
mnci install -w api github.com/pkg/errors  # go get, in apps/api
mnci uninstall -w web zod
```

`-w` takes a project's name or its directory, and can be repeated. `-D` adds a
development dependency: an npm devDependency, the Python `dev` extra, or a Go
tool (`go get -tool`). A bare `mnci install` installs everything a fresh clone
needs: `npm install`, the Python venv with every project in it, and
`go mod download`. Without `-w`, a package is refused instead of being added
at the root.

## Run tasks

```sh
mnci build                     # every project
mnci test -p web               # web and what it depends on
mnci run lint typecheck test   # several tasks
mnci run start -p web          # run an app
mnci test --affected           # only what changed since main
```

A task whose inputs did not change replays from the cache. See
[tasks and cache](tasks-and-cache.md).

## Commit and release

Commit with [Conventional Commits](https://www.conventionalcommits.org):

```sh
git commit -m "feat(sdk): add retries"
mnci release --dry-run
```

On `main`, CI runs `mnci ci release`, which publishes every package with
releasable commits. See [releasing](releasing.md).

## Check the workspace

```sh
mnci doctor          # every failed check names its remedy
mnci explain web     # how web's tasks resolved and where each value came from
```

## The interactive shell

`mnci` with no arguments, in a terminal, opens a shell with these tabs (tab and
shift+tab switch between them):

- **Menu**: the common tasks, asking one question at a time.
- **Commands**: every command; type to find one, press enter, then type its
  arguments.
- **Doctor** (inside a workspace): every check, problems first. Press enter on
  one to see its detail and remedy, and r to check again.
- **Release** (inside a workspace): what the next release would do, or why it
  is blocked. Press p to preview it and R to release.

The shell never acts by itself. Each action ends as an mnci command line,
printed before it runs: you can type the same command in a script or in CI. R,
for example, hands over to `mnci release`, which asks for confirmation and then
shows its progress. When stdin or stdout is not a terminal, or `CI` is set,
`mnci` prints its help instead.
