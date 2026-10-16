# Stacks

A **stack** is the set of plugins a project is built with. Each project
records its own, so one monorepo can hold a React app built with nothing but
Vite next to an Azure Function built with a strongly opinionated stack:
scaffold, lint rules, test library, layout rules and deploy steps.

## Plugin roles

Every plugin plays one or more roles:

| Role | Does | Per project |
|---|---|---|
| **language** | finds projects from their manifests, turns imports into graph edges, writes versions, publishes | exactly one |
| **toolchain** | how the project builds, tests and lints | any number; two may not provide the same task |
| **structure** | layout and code rules, checked as tasks | any number |
| **scaffold** | the starter files of a new project | at most one |
| **deploy** | pipeline steps and deploy targets | any number |

Any plugin may add `doctor` checks and CI steps for the projects it serves.

mnci itself knows no language: the graph, the cache, affected and releases work
the same for every project, because dependencies always come from the language
plugin. A project built with one stack can depend on a project built with
another.

A language plugin without explicit roles also brings a default toolchain and
scaffold, so `node`, `go` and `python` work on their own: a project with no
stack is built by its language plugin alone.

## Declaring a stack

Plugins publish ready-made stacks, and `mnci.toml` declares more (a stack
declared there wins over a published one of the same name):

```toml
[stacks.ts-func]
description = "Azure Function in TypeScript, the team's way"
language = "node"
toolchain = ["eslint", "vitest"]
structure = ["team-layout"]
scaffold = "azure-function"
deploy = ["azure-functions"]
```

Toolchain plugins apply in order, so a later one may replace a task an earlier
one added. mnci checks every stack against the loaded plugins: each is loaded,
plays the role it is listed under and applies to the stack's language, and no
two provide the same task (`jest` and `vitest` both provide `test`).

`mnci stacks` lists every stack, where it comes from, its plugins, what is
wrong with it if it cannot be used, and the projects using it.

## Choosing a stack

```sh
mnci add app api --stack ts-func
mnci add app web --lang node          # the default stack for node, if any
```

The project records its stack in its manifest:

| Language | Where |
|---|---|
| Node | `"mnci": { "kind": "app", "stack": "ts-func" }` in `package.json` |
| Python | `stack = "ts-func"` under `[tool.mnci]` in `pyproject.toml` |
| Go | a `// mnci:stack ts-func` line at the top of `go.mod` |

A project's stack is, in order: `[projects.<name>] stack` in `mnci.toml`, the
one its manifest records, or the default for its language:

```toml
[default_stacks]
node = "node-jest"
```

## How a project's tasks are built

1. The language plugin discovers the project and its tasks.
2. The stack's toolchain, structure and deploy plugins contribute, in order.
   Each adds tasks the project lacks, may **replace** one it has, and may make
   an existing task depend on one of its own.
3. Conventions apply to every project of their languages
   ([conventions](conventions.md)).
4. `mnci.toml` task overrides apply last, so you always have the final word.

`mnci explain <project>:<task>` shows where every field came from, for example
`from plugin vitest (stack node-vitest)`.

## Creating a project

`mnci add` asks the stack's scaffold (or, without one, the language plugin) to
generate the project and record the stack, then asks each of the stack's other
plugins to add its part: files, scripts, settings, dependencies. These only
add: a file, script, setting or dependency that already exists is kept.

## Built-in stacks

| Stack | Language | Toolchain | Published by |
|---|---|---|---|
| `node-jest` | node | `eslint` (`@mnci/eslint-config`), `jest` | `jest` |
| `node-vitest` | node | `eslint` (`@mnci/eslint-config`), `vitest` | `vitest` |

`mnci new` makes `node-jest` the default for Node. The node plugin alone
builds and type-checks with `tsc`. The toolchain plugins add the rest when a
project is created:

- `eslint` writes `eslint.config.mjs` at the root (once), adds `eslint` and
  `@mnci/eslint-config` to the root devDependencies, and a `lint` script.
- `jest` adds `jest`, `ts-jest` and `@jest/globals`, a `test` script, the
  Jest settings in `package.json`, and the starter's spec.
- `vitest` adds `vitest`, a `test` script (`vitest run`) and the starter's
  spec.

A workspace created by mnci 0.1 had ESLint and Jest built into the node
plugin. Add `[default_stacks]` with `node = "node-jest"` to its `mnci.toml` to
keep new projects the same; existing projects are unchanged.

## Adding plugins

Any executable speaking the [plugin protocol](plugins.md) is a plugin,
written in any language. Declare it, then use it in a stack:

```toml
[plugins.external.azure-function]
command = ["mnci-plugin-azure-function"]
```

A new language is a language plugin like any other. Installing plugins by name
from a registry is planned (#164).
