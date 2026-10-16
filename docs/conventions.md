# Conventions

A language plugin knows how to build, test and release a kind of project. A
**convention** adds the rules a team holds on top: how code is laid out, which
checks run, what a pipeline adds. A convention applies to the projects of the
languages it supports, in a workspace that turns it on in `smith.toml`:

```toml
[conventions.vertical-slices]
# languages = ["go"]     # default: every language the convention supports
# exclude = ["legacy"]   # projects it leaves alone
```

`smith new` turns on `vertical-slices`. A converted workspace does not get it
until you add the line, because an existing codebase rarely follows it yet.

A convention is any plugin applied workspace-wide. To give some projects a
plugin and not others, put it in their [stack](stacks.md) instead: a
structure plugin such as `vertical-slices` can be either.

| Convention | Languages | What it does |
|---|---|---|
| `vertical-slices` | Go, Python, Dart | a `slice-check` task, run before `lint` |

## How a convention combines with a project

1. The language plugin discovers each project and its tasks, and the
   project's [stack](stacks.md) adds its plugins' tasks.
2. Each configured convention adds its tasks to the projects it applies to. A
   task the project already has is kept as it is; a convention can also add a
   dependency to an existing task (`lint` depends on `slice-check`).
3. `smith.toml` overrides apply last, so they win over both.

`smith explain <project>:<task>` shows where each field came from
(`from convention vertical-slices`). A convention that answers `ci` or
`doctor` is asked about the projects it applies to, so its pipeline steps and
checks join the language plugins'. A convention that can `generate` adds its
part to every new project it applies to.

## vertical-slices

Code is organised by what it does: one folder per outcome (a slice), flat, its
files named for the role they play.

| Language | Checked by | Rules |
|---|---|---|
| Go | `slice-check` task (`smith slices check --lang go`) | a file is `<name>_<role>.go` (`_test`, and `_windows`-style platform suffixes allowed); only `doc.go` and `main.go` sit at a module's root |
| Python | `slice-check` task (`smith slices check --lang python`) | a module is `<name>_<role>.py`, a test `test_<name>.py`; only `__init__.py` and `__main__.py` sit at a package's top folder; packages are snake_case |
| Dart, Flutter | `slice-check` task (`smith slices check --lang dart`) | files and folders are lowercase_with_underscores (Effective Dart); a slice under `lib/src/` holds `<name>_<role>.dart` files and its barrel `<slice>.dart`; `lib/` holds the public libraries (and a Flutter app's `main.dart`), `bin/` the entry points; tests mirror `lib/` under `test/` as `<basename>_test.dart`, with test doubles in a `testing/` folder; generated `*.g.dart`-style files and the platform folders are not checked |
| Node | `@mnci/eslint-config`'s vertical-slice rules, in `lint` (the `eslint` toolchain plugin of the `node-jest` and `node-vitest` stacks) | the same rules, plus slice boundaries and cycles |

The roles are `handler`, `use_case`, `algorithm`, `policy`, `model`,
`contract`, `mapper`, `validator`, `repository`, `client`, `store`,
`middleware`, `error`, `config` and `enum`. No folder is named `util(s)`,
`helper(s)`, `common`, `shared`, `manager`, `processor`, `data` or `misc`.
`testdata`, `fixtures`, `vendor`, `dist`, `build` and dot-folders are not
checked.

A slice with more than 12 production files (tests, `doc.go`,
`__init__.py` and Dart barrels do not count) is reported for review, without failing: it
usually holds two outcomes and should become two slices. If it should not,
write the exception down in the project's README.

`slice-check` is cached on the project's `.go`, `.py` or `.dart` files, so it runs
again only when one of them changes, appears or disappears.
