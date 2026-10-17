# Releasing

```sh
smith release --dry-run    # the plan; changes nothing
smith release --dry-run --comment   # in a pull request build: also post the plan on it (docs/ci.md)
smith release              # asks before releasing (CI: smith ci release)
```

## What releases

Every project of kind `package`, and every app with a release section (see
[Apps](#apps)). For each, smith reads the commits since its last release tag that
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

An app releases when `smith.toml` gives it a release section:

```toml
[projects.cli.release]
asset_name = "{product}_{version}_{os}_{arch}.{ext}"
```

It is versioned by its tag (`cli@1.2.0`) only. Its plugin builds the release
assets. Today only the Go plugin builds them: it cross-compiles every platform
with `CGO_ENABLED=0`, stamps the version into `main.version` (`-ldflags -X`),
zips each binary and writes `checksums.txt`, all under `dist/release/<app>/`.
An app in another language with a release section is tagged and gets a GitHub
Release, without assets. The settings are in
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
authors, bots aside. `smith release --dry-run --notes` prints them before
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

### Linux packages

`linux_packages` builds a .deb and an .rpm from each linux archive (amd64 and
arm64), with the binary in `/usr/bin`, and attaches them to the GitHub
Release with their SHA-256 in `checksums.txt`:

```toml
[projects.cli.release]
linux_packages = ["deb", "rpm"]
```

```sh
sudo apt install ./acme_1.2.0_amd64.deb
sudo dnf install ./acme-1.2.0-1.x86_64.rpm
```

They are built before the app is tagged, so a failure stops the release, and
a resumed release attaches any the GitHub Release lacks. A prerelease
(`1.2.0-rc.1`) is versioned `1.2.0~rc.1`, which both formats order before
`1.2.0`.

## npm, PyPI, NuGet and RubyGems

An app released with assets can also be published to npm, PyPI, NuGet and
RubyGems from the same archives, so people install it with the package
manager of their language, and nothing is downloaded after install:

```toml
[projects.cli.release]
npm = ["@acme/cli", "acme"]   # npm install --global acme
pypi = "acme"                 # pipx install acme, uvx acme
nuget = "acme"                # dotnet tool install -g acme
rubygems = "acme"             # gem install acme
```

| Registry | Packages | Secret |
|---|---|---|
| npm | One package per platform, named after the first name (`@acme/cli-linux-x64`, `@acme/cli-win32-arm64`, ...), holding the binary and limited to its `os` and `cpu`. Then every listed name as a package whose `bin` runs the binary that npm installed from its optional dependencies. | `NODE_AUTH_TOKEN` |
| PyPI | One wheel per platform, the binary in its scripts, so pip, pipx and uv put it on PATH with no Python in between. One Linux wheel serves glibc and musl. | `PYPI_TOKEN` |
| NuGet | One package per platform (`acme.linux-x64`, `acme.win-arm64`, ...) holding the binary, then the .NET tool package `acme`, which points at them. `dotnet tool install` (.NET 10 SDK and later) picks the platform's package and the command runs the binary itself, with no .NET runtime. Set `NUGET_PUSH_URL` to push to another feed's v2 endpoint. | `NUGET_API_KEY` |
| RubyGems | One platform gem per platform (`x86_64-linux`, `x86_64-linux-musl`, `arm64-darwin`, `x64-mingw-ucrt`, ...), holding the binary and a Ruby launcher that runs it; RubyGems installs the gem matching the machine. | `GEM_HOST_API_KEY` |

They are published after the archives are built and before the tag, platform
packages before the packages that depend on them. A version a registry
already has is skipped, so a re-run completes a release. On GitHub Actions
npm packages carry [provenance](https://docs.npmjs.com/generating-provenance-statements):
the generated workflow gets `id-token: write` for it. Publishing to npm needs
the `npm` CLI on the runner (GitHub's and Azure's hosted images have it). A
prerelease goes to npm's `next` tag, to NuGet as is, to PyPI as `1.0.0rc1`
(only `-alpha.N`, `-beta.N` and `-rc.N` have a Python spelling), and to
RubyGems as `1.0.0.pre.rc.1`. Native apps are not published
this way.

## Tags

| Language | Tag | Why |
|---|---|---|
| every language but Go, and every app | `<name>@<version>` | override with `tag_prefix` |
| Go packages | `<dir>/v<version>` (e.g. `packages/gokit/v0.1.0`) | how `go get` finds a nested module's version |

The model is **tag-only**: nothing is committed back to the branch. For each
package, in dependency order, smith:

1. writes the version into the manifest (keeping its formatting), and the new
   versions of workspace dependencies into its dependency ranges
2. builds the package
3. publishes it — skipping a version the registry already has
4. restores the manifest
5. tags the commit and pushes the tag

## Tag checks

Versions come from tags, so before showing a plan, even for `--dry-run`,
`smith release` fetches the tags from the release remote and checks them for
every releasable project. A failed fetch is a warning. Each of these stops the
release, because the plan's versions would be wrong:

| Problem | Why it matters | Fix |
|---|---|---|
| The project's newest tag is under another name (its package name, scoped or not), not its tag prefix | the release would start from scratch, or from an older version | set `[projects.<name>] tag_prefix`, or tag the same commit under the current prefix |
| The registry has a version, but the project has no tag | the release would start from the manifest | `git fetch --tags`; if the tag really does not exist, tag the commit that published it |
| The newest tag is lower than the registry's version | the release would propose a version the registry refuses | tag the commit that published the registry's version |

The registry is asked through the project's plugin (`npm view` for Node, the
PEP 503 index for Python, and likewise crates.io, NuGet, RubyGems, pub.dev,
Packagist and the Maven repository; a Go module's tags are its registry, and
C++ and Bicep projects have none). An unreachable
registry is a warning, not a stop. `smith doctor` runs the same checks without
fetching.

## When something fails

The release stops at the failing package; everything before it is tagged and
pushed. Re-run it: tagged packages have no new commits and are kept, and a
package that was published but not tagged is skipped by the registry check and
then tagged. Tags left local by an interrupted push are pushed by the next run.

Before a real release smith refuses a shallow clone, a branch other than
`[release] branch`, uncommitted changes, and a missing publish credential —
before anything is tagged.

## The report

Every run writes `.smith/release-report.json` (also printed by `--json`) and, on
GitHub Actions, a summary table:

```json
{ "dryRun": false, "results": [
  { "project": "sdk", "current": "0.1.0", "next": "0.2.0", "tag": "sdk@0.2.0",
    "reason": "2 releasing commit(s)", "status": "released",
    "published": true, "registry": "https://registry.npmjs.org/", "pushed": true } ],
  "unchanged": ["gokit"] }
```

There are no changelog files: the release notes go to the GitHub Release.
