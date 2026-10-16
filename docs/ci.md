# CI pipelines

mnci generates the pipeline from what is in the workspace and keeps it in step:
`mnci new` writes it, every `mnci add` regenerates it, and `mnci ci generate`
does it on demand. `mnci doctor` (and `mnci ci generate --check`) fails when the
committed pipeline is out of date.

## How it is built

Each plugin contributes, for its own projects:

- **toolchains** — the runtimes to install, at versions the plugin pins. The
  plugin says how: a GitHub action and an Azure task with their inputs
  (`actions/setup-node@v4`, `UsePythonVersion@0`, …), or a shell command run
  on either provider. mnci itself knows no runtime, so a plugin for a new
  language installs it the same way; a runtime no plugin says how to install
  is left as a comment in the pipeline, to fill through the `before-setup`
  slot.
- **setup** — e.g. `npm ci`, the Python venv and editable installs, `go mod download`
- **audit** — `npm audit` (blocking), `pip-audit` and `govulncheck` (reported)
- **credentials** — the secrets publishing needs (`NODE_AUTH_TOKEN`, `PYPI_TOKEN`)

mnci merges them into one provider-neutral model, and a renderer per provider
writes the YAML. Both providers carry the same phases, in the same order.

## The phases

The YAML is thin: each step calls `mnci ci <phase>`, so what CI runs is what you
can run locally.

| Phase | Does |
|---|---|
| `mnci ci setup` | every setup step the plugins contributed |
| `mnci ci audit` | dependency vulnerability checks; non-blocking ones warn |
| `mnci ci verify` | `lint`, `typecheck`, `test`, `build` — affected projects only on a pull request, everything otherwise |
| `mnci release --dry-run --comment` | on a pull request: the release preview, in the job summary and as a comment on the pull request |
| `mnci ci pack` | on a push to the release branch: every project's `package` task into `dist/drop`, kept as the run's `drop` artifact |
| `mnci ci release` | on a push to the release branch: version, publish and tag |

On GitHub Actions and Azure Pipelines output is grouped, failures are annotated,
and each phase adds a line (passed or failed, and how long it took) and the
release its report to the job summary.

## Packing

`mnci ci pack` runs the `package` task of every project that has one and
leaves what they make in `dist/drop`, which the pipeline keeps as the `drop`
artifact (GitHub `actions/upload-artifact`, Azure `publish`). The steps are in
the pipeline only when some project has a `package` task. The built-in
plugins give every app one:

| App | `package` makes |
|---|---|
| Node | `npm pack` of the app, after its build: `dist/drop/<scope>-<name>-<version>.tgz` |
| Go | the binary: `dist/drop/<name>/<name>` |
| Python | the app's wheel: `dist/drop/<name>-<version>-py3-none-any.whl` |

A stack's plugin, or `[projects.<name>.tasks.package]` in `mnci.toml`, can
replace it, for example with a zip a deploy step expects.

## The release preview comment

On a pull request the preview is posted as one comment, which later runs of
the same pull request update instead of adding more: an issue comment on
GitHub (the workflow grants `pull-requests: write` and passes
`GITHUB_TOKEN`), a closed thread on Azure Repos (the step passes
`SYSTEM_ACCESSTOKEN: $(System.AccessToken)`; the build service needs
"Contribute to pull requests"). A pull request from a fork gets a read-only
token, so its preview stays in the log and the job summary; failing to post
is a warning, never a failure.

## Pinned versions and the dev container

Nothing installs "latest". Each plugin pins what it installs: Node.js 24,
Python 3.12 and its tools (`requirements-dev.txt`), Go at the version
`go.work` names and golangci-lint 2.14.0; actions are pinned to a major.
Generated pipelines install mnci at the version that generated them.

With `[ci] devcontainer = true` (which `mnci new` sets), mnci also writes
`.devcontainer/devcontainer.json`: a base image plus one dev container
feature per toolchain, at the versions CI installs, and a post-create command
that installs the same mnci and runs `mnci ci setup`. Like the pipelines, it
is rewritten by `mnci add` and `mnci ci generate`, and `ci generate --check`
and doctor report it when it is out of date.

`mnci doctor` compares the tools on your machine with the pins: an older Go
than `go.work` requires fails; a different Node.js major, venv Python or
golangci-lint warns, since results may then differ from CI's.

## Installing mnci in CI

Generated pipelines install the mnci version that generated them with the
install script (`MNCI_VERSION`), so a pipeline does not change behaviour until
you regenerate it with a newer mnci.

## Secrets

| Secret | Used for |
|---|---|
| `NODE_AUTH_TOKEN` | `npm publish` (the generated `.npmrc` reads it) |
| `PYPI_TOKEN` | twine upload as `__token__` (or set `TWINE_USERNAME`/`TWINE_PASSWORD` and `TWINE_REPOSITORY_URL` for another index) |
| `GITHUB_TOKEN` | GitHub Releases and their assets, with `[release] github_releases`; GitHub Actions provides it |

Go packages are published by their tag and need no secret. On GitHub, the job
has `contents: write` so it can push tags; on Azure the checkout keeps its
credentials (`persistCredentials`).

## Your own steps

Put team steps between the slot markers; regeneration keeps them:

```yaml
      # mnci:slot after-verify
      - name: End-to-end tests
        run: npm run e2e
      # mnci:slot-end after-verify
```

Slots: `before-setup`, `after-verify`.
