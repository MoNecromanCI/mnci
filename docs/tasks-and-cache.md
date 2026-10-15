# Tasks, cache and affected

## Tasks

A task is a command a plugin infers for a project: `build`, `test`, `lint`,
`typecheck`, `start` and any npm script. Each has:

| Field | Meaning |
|---|---|
| `command` | argv, run without a shell, from the project root; `{workspaceRoot}`, `{projectRoot}` and `{mnci}` (the running mnci binary) are expanded |
| `dependsOn` | `build` — this project's build first; `^build` — the build of every project this one imports first (reaching through projects without a build); `web:build` — a specific task |
| `inputs` | globs relative to the project, `{workspaceRoot}/...` for root files, `!` to exclude; default: every file of the project |
| `outputs` | directories restored on a cache hit (`dist`) |
| `cache` | whether results are cached |

```sh
mnci run <task>... [-p <projects>] [--affected] [--parallel N] [--bail] [--no-cache] [--json]
mnci build | test | lint | typecheck [project...]
```

Project selection (`-p`, comma-separated or repeated): a name, a glob over names
or directories (`apps/*`), `kind:package`, `lang:go`, `tag:x`, and `!` to
exclude.

Every task runs with the workspace's Python environment (`.venv`) first on
`PATH`, the way npm puts `node_modules/.bin` first, so `python` and the tools
installed in the venv are the workspace's on every OS.

Tasks run in parallel (one per CPU by default) in dependency order. A failed task
skips the tasks that depend on it; independent tasks keep going unless `--bail`.
Ctrl+C stops every process tree mnci started — on Windows through a Job Object,
elsewhere through a process group.

## Cache

A task's **key** is a SHA-256 over a versioned, canonical encoding of:

- the task id and its command (before `{workspaceRoot}` and `{mnci}` are expanded, so keys do
  not depend on where the repository is checked out)
- the declared env
- the content of every input file
- the sources of every project it depends on, transitively
- the keys of the tasks it depends on
- the OS/architecture and the mnci and plugin versions

On a hit, mnci replaces the task's output directories with the cached files and
replays its log. On a miss it clears the outputs, runs the task, and stores the
result if it succeeded.

The store is a directory (`.mnci/cache/v1`) split like Bazel's remote cache:
`ac/` holds one small JSON entry per key; `cas/` holds blobs named by their
SHA-256, shared between entries. Writes land by rename, so an interrupted run
never leaves a half-written entry; an entry with a missing blob is a miss. When
the cache grows past `max_size_mb`, the least recently used entries go first.

Everything goes through a `Storage` interface (`Has/Get/Put/Delete/Walk/Touch`)
with a conformance test, so a remote backend (HTTP, S3, Azure Blob — #42) is an
addition, not a change.

## Affected

```sh
mnci affected [--base main]
mnci test --affected
```

The affected projects are those owning a changed file (since the merge base
with the base branch, plus uncommitted and untracked files) and every project
that depends on them. The base is `--base`, else the pull request target
(`GITHUB_BASE_REF`, `SYSTEM_PULLREQUEST_TARGETBRANCH`, `MNCI_BASE`), else the
release branch.

When mnci cannot be sure it affects **everything**: no base, an unknown base, no
common ancestor (a shallow clone), or a changed workspace-level file such as
`mnci.toml` or a lockfile. Documentation (`*.md`, `docs/**`) never affects a
project. Too wide costs minutes; too narrow reports green for code nothing
checked.
