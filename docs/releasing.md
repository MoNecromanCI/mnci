# Releasing

```sh
mnci release --dry-run    # the plan; changes nothing
mnci release --dry-run --comment   # in a pull request build: also post the plan on it (docs/ci.md)
mnci release              # asks before releasing (CI: mnci ci release)
```

## What releases

Every project of kind `package`, and every app with a release section (see
[Apps](#apps)). For each, mnci reads the commits since its last release tag that
touched its directory, or the directory of an internal `lib` it imports (a lib
ships inside what uses it, so a fix there releases its users):

| Commit | Bump |
|---|---|
| `feat:` | minor |
| `fix:`, `perf:` | patch |
| `!` after the type, or a `BREAKING CHANGE:` footer | major (minor before 1.0) |
| `docs`, `chore`, `test`, `ci`, `refactor`, `build`, `style` | nothing |

The first release starts from the manifest's version (`0.0.0`), so a first
`feat` gives `0.1.0`. A package whose dependency releases is released too (at
least a patch), so its published dependency range stays correct.

To force a version, commit a directive anywhere in the repository:

```
version(sdk)[1.0.0]: first stable release
```

## Apps

An app releases when `mnci.toml` gives it a release section:

```toml
[projects.cli.release]
asset_name = "{product}_{version}_{os}_{arch}.{ext}"
```

It is versioned by its tag (`cli@1.2.0`) only. Its plugin builds the release
assets: the Go plugin cross-compiles every platform with `CGO_ENABLED=0`,
stamps the version into `main.version` (`-ldflags -X`), zips each binary and
writes `checksums.txt`, all under `dist/release/<app>/`. The settings are in
[Configuration](configuration.md#a-projects-release-settings).

## GitHub Releases

With `[release] github_releases = true`, every released tag gets a GitHub
Release named `<project> <version>`, and an app's assets are attached. It
needs `GITHUB_TOKEN` (or `GH_TOKEN`) with `contents: write`, which the
generated GitHub pipeline passes. A re-run completes a release that is missing
assets and never duplicates one.

The notes come from the releasing commits. Breaking changes are listed first,
with the text of their `BREAKING CHANGE:` footer. Then come Features, Fixes,
Performance and Reverts, each entry with its scope and short hash. A release
caused only by a dependency says so. The notes end by thanking the commits'
authors, bots aside. `mnci release --dry-run --notes` prints them before
anything is released.

## Package managers

An app released with assets can also get the manifests package managers
install it from, generated from the archives and their checksums, and attached
to the GitHub Release beside them:

```toml
[projects.cli.release]
package_managers = ["homebrew", "scoop", "winget"]
description = "What the app does"
license = "MIT"
homepage = "https://cli.acme.dev"   # default: the GitHub repository
```

| Manager | Files | Where they go |
|---|---|---|
| Homebrew | `<product>.rb` (macOS and Linux, Apple silicon and Intel) | `Formula/` in your tap |
| Scoop | `<product>.json` (64-bit and arm64) | your bucket |
| winget | `<Owner>.<product>.yaml`, `.installer.yaml`, `.locale.en-US.yaml` (a portable zip) | a pull request to `microsoft/winget-pkgs` |

They point at the release's own archives with their SHA-256, so each one can
be copied into a tap, a bucket or winget-pkgs as it is. A manager whose
platforms the release has no archive for is skipped, and the release says so.
Native apps attach one archive per OS from separate jobs and get no manifests.

## Tags

| Language | Tag | Why |
|---|---|---|
| Node, Python, apps | `<name>@<version>` | override with `tag_prefix` |
| Go | `<dir>/v<version>` (e.g. `packages/gokit/v0.1.0`) | how `go get` finds a nested module's version |

The model is **tag-only**: nothing is committed back to the branch. For each
package, in dependency order, mnci:

1. writes the version into the manifest (keeping its formatting), and the new
   versions of workspace dependencies into its dependency ranges
2. builds the package
3. publishes it — skipping a version the registry already has
4. restores the manifest
5. tags the commit and pushes the tag

## Tag checks

Versions come from tags, so before showing a plan, even for `--dry-run`,
`mnci release` fetches the tags from the release remote and checks them for
every releasable project. A failed fetch is a warning. Each of these stops the
release, because the plan's versions would be wrong:

| Problem | Why it matters | Fix |
|---|---|---|
| The project's newest tag is under another name (its package name, scoped or not), not its tag prefix | the release would start from scratch, or from an older version | set `[projects.<name>] tag_prefix`, or tag the same commit under the current prefix |
| The registry has a version, but the project has no tag | the release would start from the manifest | `git fetch --tags`; if the tag really does not exist, tag the commit that published it |
| The newest tag is lower than the registry's version | the release would propose a version the registry refuses | tag the commit that published the registry's version |

The registry is asked through the project's plugin (`npm view` for Node, the
PEP 503 index for Python; a Go module's tags are its registry). An unreachable
registry is a warning, not a stop. `mnci doctor` runs the same checks without
fetching.

## When something fails

The release stops at the failing package; everything before it is tagged and
pushed. Re-run it: tagged packages have no new commits and are kept, and a
package that was published but not tagged is skipped by the registry check and
then tagged. Tags left local by an interrupted push are pushed by the next run.

Before a real release mnci refuses a shallow clone, a branch other than
`[release] branch`, uncommitted changes, and a missing publish credential —
before anything is tagged.

## The report

Every run writes `.mnci/release-report.json` (also printed by `--json`) and, on
GitHub Actions, a summary table:

```json
{ "dryRun": false, "results": [
  { "project": "sdk", "current": "0.1.0", "next": "0.2.0", "tag": "sdk@0.2.0",
    "reason": "2 releasing commit(s)", "status": "released",
    "published": true, "registry": "https://registry.npmjs.org/", "pushed": true } ],
  "unchanged": ["gokit"] }
```

Changelog files are planned (#64).
