# CI pipelines

mnci generates the pipeline from what is in the workspace and keeps it in step:
`mnci new` writes it, every `mnci add` regenerates it, and `mnci ci generate`
does it on demand. `mnci doctor` (and `mnci ci generate --check`) fails when the
committed pipeline is out of date.

## How it is built

Each plugin contributes, for its own projects:

- **toolchains** — rendered as provider-native setup steps (`actions/setup-node`,
  `UsePythonVersion@0`, …), at versions the plugin pins
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
| `mnci release --dry-run` | on a pull request: the release preview, in the job summary |
| `mnci ci release` | on a push to the release branch: version, publish and tag |

On GitHub Actions and Azure Pipelines output is grouped, failures are annotated,
and the release report is appended to the job summary.

## Installing mnci in CI

Generated pipelines install the mnci version that generated them with the
install script (`MNCI_VERSION`), so a pipeline does not change behaviour until
you regenerate it with a newer mnci.

## Secrets

| Secret | Used for |
|---|---|
| `NODE_AUTH_TOKEN` | `npm publish` (the generated `.npmrc` reads it) |
| `PYPI_TOKEN` | twine upload as `__token__` (or set `TWINE_USERNAME`/`TWINE_PASSWORD` and `TWINE_REPOSITORY_URL` for another index) |

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
