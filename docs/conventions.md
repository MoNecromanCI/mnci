# Conventions

A language plugin knows how to build, test and release a kind of project. A
**convention** adds the rules a team holds on top: how code is laid out, which
checks run, what a pipeline adds. A convention applies to the projects of the
languages it supports, in a workspace that turns it on in `mnci.toml`:

```toml
[conventions.vertical-slices]
# languages = ["go"]     # default: every language the convention supports
# exclude = ["legacy"]   # projects it leaves alone
```

`mnci new` turns on `vertical-slices`, `eslint` and `jest`. A converted
workspace gets none of them until you add the lines: an existing codebase
rarely follows vertical slices yet, and it already has its own lint and test
setup.

| Convention | Languages | What it does |
|---|---|---|
| `vertical-slices` | Go, Python | a `slice-check` task, run before `lint` |
| `eslint` | Node | new projects lint with ESLint and `@mnci/eslint-config` |
| `jest` | Node | new projects test with Jest and `ts-jest` |
| `vitest` | Node | new projects test with Vitest |

## How a convention combines with a project

A convention can do two things: add tasks to every project it applies to
(`vertical-slices`), and set a project up when `mnci add` generates it
(`eslint`, `jest`, `vitest`).

### Tasks

1. The language plugin discovers each project and its tasks.
2. Each configured convention adds its tasks to the projects it applies to. A
   task the project already has is kept as it is; a convention can also add a
   dependency to an existing task (`lint` depends on `slice-check`).
3. `mnci.toml` overrides apply last, so they win over both.

`mnci explain <project>:<task>` shows where each field came from
(`from convention vertical-slices`). A convention that answers `ci` or
`doctor` is asked about the projects it applies to, so its pipeline steps and
checks join the language plugins'.

### New projects

`mnci add` asks the language plugin to generate the project, then asks each
configured convention that applies to it to add its part. A convention only
adds: a file, script, field or dependency that already exists is kept.

The node plugin generates a TypeScript project that builds and type-checks
with `tsc`. The Node conventions add the rest:

- `eslint` writes `eslint.config.mjs` at the root (once), adds `eslint` and
  `@mnci/eslint-config` to the root devDependencies, and a `lint` script.
- `jest` adds `jest`, `ts-jest` and `@jest/globals`, a `test` script, the
  Jest settings in `package.json`, and the starter's spec.
- `vitest` adds `vitest`, a `test` script (`vitest run`) and the starter's
  spec.

Configure one test runner. Switching changes only projects added after the
switch; existing projects keep the runner they were set up with, so a
workspace can move from Jest to Vitest one project at a time.

A workspace created before these conventions existed (mnci 0.1 and earlier)
had Jest and ESLint built into the node plugin. Add `[conventions.eslint]`
and `[conventions.jest]` to its `mnci.toml` to keep new projects the same.

## vertical-slices

Code is organised by what it does: one folder per outcome (a slice), flat, its
files named for the role they play.

| Language | Checked by | Rules |
|---|---|---|
| Go | `slice-check` task (`mnci slices check --lang go`) | a file is `<name>_<role>.go` (`_test`, and `_windows`-style platform suffixes allowed); only `doc.go` and `main.go` sit at a module's root |
| Python | `slice-check` task (`mnci slices check --lang python`) | a module is `<name>_<role>.py`, a test `test_<name>.py`; only `__init__.py` and `__main__.py` sit at a package's top folder; packages are snake_case |
| Node | `@mnci/eslint-config`'s vertical-slice rules, in `lint` (the `eslint` convention) | the same rules, plus slice boundaries and cycles |

The roles are `handler`, `use_case`, `algorithm`, `policy`, `model`,
`contract`, `mapper`, `validator`, `repository`, `client`, `store`,
`middleware`, `error`, `config` and `enum`. No folder is named `util(s)`,
`helper(s)`, `common`, `shared`, `manager`, `processor`, `data` or `misc`.
`testdata`, `fixtures`, `vendor`, `dist`, `build` and dot-folders are not
checked.

A slice with more than 12 production files (tests, `doc.go` and
`__init__.py` do not count) is reported for review, without failing: it
usually holds two outcomes and should become two slices. If it should not,
write the exception down in the project's README.

`slice-check` is cached on the project's `.go` or `.py` files, so it runs
again only when one of them changes, appears or disappears.
