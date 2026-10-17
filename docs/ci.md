# CI pipelines

smith generates the pipeline from what is in the workspace and keeps it in step:
`smith new` writes it, every `smith add` regenerates it, and `smith ci generate`
does it on demand. `smith doctor` (and `smith ci generate --check`) fails when the
committed pipeline is out of date.

## How it is built

Each plugin contributes, for its own projects:

- **toolchains** — the runtimes to install, at versions the plugin pins. The
  plugin says how: a GitHub action and an Azure task with their inputs
  (`actions/setup-node@v4`, `UsePythonVersion@0`, …), and a shell command for
  everything else (GitLab installs every toolchain this way). smith itself knows no runtime, so a plugin for a new
  language installs it the same way; a runtime no plugin says how to install
  is left as a comment in the pipeline, to fill through the `before-setup`
  slot.
- **setup** — e.g. `npm ci`, the Python venv and editable installs, `go mod download`
- **audit** — `npm audit` (blocking), `pip-audit` and `govulncheck` (reported)
- **credentials** — the secrets publishing needs (`NODE_AUTH_TOKEN`, `PYPI_TOKEN`, `NUGET_API_KEY`)

smith merges them into one provider-neutral model, and a renderer per provider
writes the YAML. Every provider carries the same phases, in the same order.

## GitLab CI

With `"gitlab"` in `[ci] providers`, smith writes `.gitlab-ci.yml`. It has
one job, run for merge requests and on pushes to the release branch, on a
`debian:bookworm` image. Each toolchain installs through its plugin's shell
command: Node.js from NodeSource, Python through uv at the pinned version, Go
from its release tarball. Then come the same phases as GitHub and Azure, with
the same slots.

Two masked CI/CD variables complete it. `GITLAB_TOKEN` is a project access
token with the `api` scope, for the release preview on merge requests, and
`write_repository`, to push release tags. Then add the publishing secrets the
file lists at its end. Native projects build per OS in the GitHub and Azure
pipelines only: on GitLab, add a job per runner OS that runs
`smith ci native`.

## The phases

The YAML is thin: each step calls `smith ci <phase>`, so what CI runs is what you
can run locally.

| Phase | Does |
|---|---|
| `smith ci setup` | every setup step the plugins contributed |
| `smith ci audit` | dependency vulnerability checks; non-blocking ones warn |
| `smith ci verify` | `lint`, `typecheck`, `test`, `build` — affected projects only on a pull request, everything otherwise |
| `smith release --dry-run --comment` | on a pull request: the release preview, in the job summary and as a comment on the pull request |
| `smith ci pack` | on a push to the release branch: every project's `package` task into `dist/drop`, kept as the run's `drop` artifact |
| `smith ci native` | in a job per OS, after the ci job: lint, test, build and package the projects tagged native |
| `smith ci release` | on a push to the release branch: version, publish and tag |

On GitHub Actions and Azure Pipelines output is grouped, failures are annotated,
and each phase adds a line (passed or failed, and how long it took) and the
release its report to the job summary.

## Packing

`smith ci pack` runs the `package` task of every project that has one and
leaves what they make in `dist/drop`, which the pipeline keeps as the `drop`
artifact (GitHub `actions/upload-artifact`, Azure `publish`). The steps are in
the pipeline only when some project has a `package` task. The built-in
plugins give every app one:

| App | `package` makes |
|---|---|
| Node | `npm pack` of the app, after its build: `dist/drop/<scope>-<name>-<version>.tgz` |
| Go | the binary: `dist/drop/<name>/<name>` (a plain `smith build` writes it to `dist/<project dir>/`) |
| Python | the app's wheel: `dist/drop/<name>-<version>-py3-none-any.whl` |

A stack's plugin, or `[projects.<name>.tasks.package]` in `smith.toml`, can
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
Generated pipelines install smith at the version that generated them.

With `[ci] devcontainer = true` (which `smith new` sets), smith also writes
`.devcontainer/devcontainer.json`: a base image plus one dev container
feature per toolchain, at the versions CI installs, and a post-create command
that installs the same smith and runs `smith ci setup`. Like the pipelines, it
is rewritten by `smith add` and `smith ci generate`, and `ci generate --check`
and doctor report it when it is out of date.

`smith doctor` compares the tools on your machine with the pins: an older Go
than `go.work` requires fails; a different Node.js major, venv Python or
golangci-lint warns, since results may then differ from CI's.

## Native projects

A project tagged `native` must be built on each operating system rather than
cross-compiled: a Go app using cgo, marked with a `// smith:native` line at the
top of its `go.mod` (its tasks then build with `CGO_ENABLED=1`). `smith ci
verify` and `smith ci pack` leave native projects out; the generated pipelines
add a `native` job that runs after the ci job on each OS and calls `smith ci
native`, which lints, tests, builds and packages them there and keeps
`dist/drop` as the `native-<os>` artifact.

A native app that releases (`[projects.<name>.release]`) gets no
cross-compiled assets from the release: on a push that released it, each
native job builds the app for its own platform and attaches the archive to
the app's GitHub Release (GitHub Actions; the job needs `GITHUB_TOKEN`).

```toml
[ci]
native_runners = ["ubuntu-latest", "windows-latest", "macos-latest"]  # the default
native_linux_packages = ["libsqlite3-dev"]   # apt packages installed before building on Linux
```

On Azure Pipelines a workspace with a native project gets a `jobs:` pipeline:
the ci job, and the native job with a matrix of runner images. Steps kept in
the `smith:slot` markers move with the job.

## Installing smith in CI

Generated pipelines install the smith version that generated them with the
install script (`SMITH_VERSION`), so a pipeline does not change behaviour until
you regenerate it with a newer smith.

## Secrets

| Secret | Used for |
|---|---|
| `NODE_AUTH_TOKEN` | `npm publish` (the generated `.npmrc` reads it), and an app's npm packages |
| `PYPI_TOKEN` | twine upload as `__token__` (or set `TWINE_USERNAME`/`TWINE_PASSWORD` and `TWINE_REPOSITORY_URL` for another index), and an app's wheels |
| `NUGET_API_KEY` | an app's .NET tool packages; a nuget.org API key that can push the tool's package and `<name>.*` |
| `GEM_HOST_API_KEY` | an app's platform gems, pushed to rubygems.org (or `RUBYGEMS_HOST`) |
| `GITHUB_TOKEN` | GitHub Releases and their assets, with `[release] github_releases`; GitHub Actions provides it |

Go packages are published by their tag and need no secret. On GitHub, the job
has `contents: write` so it can push tags; on Azure the checkout keeps its
credentials (`persistCredentials`).

## Ready-made pipelines

Beside the generated CI, `smith pipeline` adds a pipeline that does one thing:

| Template | Providers | What it does |
|---|---|---|
| `e2e` | GitHub, Azure | runs every project's `e2e` task, which `smith ci verify` leaves out |
| `package-zip` | GitHub, Azure | packs the projects (`smith ci pack`) and keeps `dist/drop` as the `drop` artifact |
| `deploy-pages --project <app>` | GitHub | builds the app and deploys its build output to GitHub Pages |
| `deploy-azure-function --project <app>` | GitHub, Azure | builds the app and deploys it to an Azure Function App (OIDC on GitHub, a service connection on Azure) |

```sh
smith pipeline                          # list them
smith pipeline e2e                      # for the workspace's providers; --ci github|azure|both
smith pipeline deploy-pages --project web
```

Each starts the way the generated jobs do (the toolchains, smith at the same
version, `smith ci setup`) and goes to `.github/workflows/<name>.yml` or
`azure-pipelines/<name>.yml`. The file is written once and yours from then
on: `smith upgrade` never rewrites it, and adding it again keeps it unless
`--force`.

## Your own steps

Put team steps between the slot markers; regeneration keeps them:

```yaml
      # smith:slot after-verify
      - name: End-to-end tests
        run: npm run e2e
      # smith:slot-end after-verify
```

Slots: `before-setup`, `after-verify`.
