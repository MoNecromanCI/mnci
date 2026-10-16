# Getting started

## Prerequisites

- **git**, always.
- The toolchain of each language you use: **Node 24** (with npm), **Go 1.26**,
  **Python 3.12+**, **Rust** (with clippy and rustfmt), the **.NET 10 SDK**,
  **Ruby 3.2+** with Bundler,
  **PHP 8.2+** with Composer,
  **CMake 3.25+** with Ninja, a C++ compiler and cppcheck,
  **Dart** (or Flutter, which carries it; installed outside the workspace).
  mnci tells you what is missing.

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
`--ci github,azure` (or `gitlab`).

## Or adopt an existing repository

A repository with npm workspaces, Go modules or Python projects, but no
`mnci.toml`, can be adopted (an Nx workspace is `mnci convert`'s):

```sh
mnci adopt                 # read-only: what adopting involves
mnci adopt --tags --init   # take the steps
```

The report lists the projects mnci would see, then each blocker and warning
with the step that clears it. For example, a project whose release tags were
made under another name (`core@1.2.0` where mnci expects `libs/core/v1.2.0`)
would otherwise start its next release from scratch. A pipeline of yours sits
at a path mnci's uses. Root `package.json` runtime dependencies belong to the
projects that import them. `mnci adopt` exits non-zero while something blocks.

Both steps need a clean git tree. `--tags` creates the missing tags locally, on
the same commits; pushing them is yours. `--init` writes `mnci.toml` (the
name comes from the directory, the npm scope from the root `package.json`,
and the CI providers are the ones the repository already uses) and the root
files mnci needs, keeping yours and adding `.mnci/` to your `.gitignore`. It
keeps your pipeline as `<file>.pre-mnci`, generates mnci's for the projects
it found, and commits it all as `chore: adopt mnci`.

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
mnci graph --diff v1.4.0           # how the graph changed since a revision
```

Every format shows the same projects; `--focus` and `--affected` narrow all
of them the same way. `--diff` discovers the graph at another revision in a
temporary git worktree, leaving yours untouched. It lists projects and imports
added (`+`) and removed (`-`), and kinds, stacks, plugins and directories that
changed (`~`), as text or `--format json`.

Projects that import each other in a cycle cannot be ordered. `mnci graph`
warns with each cycle's path (`ui → core → ui`), and `mnci doctor` fails with
the same paths and how to break them.

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

When projects declare the same dependency at different versions, `mnci sync`
makes them agree, language by language. The target is the version installed,
with the operator most declarations use, or else the highest version declared.
Sync then installs, so lockfiles and the venv follow. `mnci sync --check`
only reports, and fails, which makes it a CI check. Peer dependencies, the
workspace's own packages and specs mnci cannot safely rewrite (URLs, git refs,
compound ranges) are left alone, and the ones that disagree are reported. Go
is not compared: a `go.work` builds every module with the highest version any
of them requires.

`mnci up` lists every declared dependency with a newer release, grouped as
patch, minor and major, with the projects that declare each one:

```sh
mnci up            # pick in the terminal (patch and minor start selected)
mnci up zod jest   # exactly these, no prompt
mnci up --yes      # everything
mnci up --check    # report only (the default without a terminal)
```

Each chosen declaration keeps its operator (`^`, `~`, `>=`…) and moves to the
latest version, then the workspace is installed again. Npm is asked through
`npm view`, so `.npmrc` applies. Python uses the index pip uses
(`PIP_INDEX_URL`, else PyPI). A final section compares the tools mnci pins
(Node.js LTS, Python, the Go version in `go.work`, golangci-lint) with their
newest releases.

## Run tasks

```sh
mnci build                     # every project
mnci test -p web               # web and what it depends on
mnci run lint typecheck test   # several tasks
mnci run start -p web          # run an app
mnci dev web api               # start several apps together; Ctrl+C stops them all
mnci test --affected           # only what changed since main
```

`mnci dev` builds what the apps use, then runs every `start` task at once,
each line prefixed with its project. With no names it lets you pick from the
apps that have a `start` task; `--all` starts every one of them.

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
mnci doctor --fix    # show the fixes doctor can make as diffs, apply the safe ones
mnci explain web     # how web's tasks resolved and where each value came from
```

`--fix` shows every fix as a unified diff before anything changes, then asks
for confirmation (`--yes` skips the question; without a terminal it only
shows). Only safe fixes are applied. A safe fix restores what mnci or a plugin
owns without losing anything of yours: a `go.work` that lists exactly the
workspace's Go modules, `.mnci/` in `.gitignore`, or pipelines regenerated with
your slot contents kept. A fix that could lose something is shown, never
applied.

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
- **Settings** (inside a workspace): the settings of `mnci.toml`, with their
  values and what they do. Enter edits one (or flips a yes/no setting), u
  returns it to its default. A value the schema rejects is refused on the
  spot. The change is made by `mnci config set` or `unset`, which keep the
  file's comments.

The shell never acts by itself. Each action ends as an mnci command line,
printed before it runs: you can type the same command in a script or in CI. R,
for example, hands over to `mnci release`, which asks for confirmation and then
shows its progress. When stdin or stdout is not a terminal, or `CI` is set,
`mnci` prints its help instead.

## Upgrade mnci

After installing a newer mnci, run it once in the workspace:

```sh
mnci upgrade           # migrate mnci.toml, regenerate the pipelines
mnci upgrade --check   # change nothing; fail when something would change (CI)
```

`mnci.toml` is migrated to the current schema with its comments and layout
kept. The files mnci owns (the pipelines, which start with "Generated by
mnci") are regenerated, with your steps between `mnci:slot` markers kept, and
one the workspace no longer asks for (a provider dropped from `[ci] providers`)
is deleted. Files mnci wrote once, like `README.md` and `.gitignore`, and your
own files are never touched. It also says what an older workspace may want,
for example a default stack for Node.
