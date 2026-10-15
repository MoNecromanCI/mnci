# Releasing

```sh
mnci release --dry-run    # the plan; changes nothing
mnci release              # asks before releasing (CI: mnci ci release)
```

## What releases

Only projects of kind `package`. For each, mnci reads the commits that touched
its directory since its last release tag:

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

## Tags

| Language | Tag | Why |
|---|---|---|
| Node, Python | `<name>@<version>` | |
| Go | `<dir>/v<version>` (e.g. `packages/gokit/v0.1.0`) | how `go get` finds a nested module's version |

The model is **tag-only**: nothing is committed back to the branch. For each
package, in dependency order, mnci:

1. writes the version into the manifest (keeping its formatting), and the new
   versions of workspace dependencies into its dependency ranges
2. builds the package
3. publishes it — skipping a version the registry already has
4. restores the manifest
5. tags the commit and pushes the tag

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

Changelogs and GitHub Releases per package are planned (#64).
