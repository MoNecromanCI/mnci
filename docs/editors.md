# Editors

Editors and tools build on four commands whose `--json` output is a stable
contract. They keep the shapes the mnci VS Code extension reads, which was
written against the TypeScript mnci.

| Command | Prints |
|---|---|
| `mnci commands --json` | every command: `name`, `aliases`, `group` (`workspace`, `projects`, `dependencies`, `pipeline`, `inspect`; an editor offers everything outside `inspect`), `description`, `arguments` (`name`, `description`, `required`, `variadic`) and `options` (`name`, `flags`, `short`, `description`, `takesValue`: `none`/`required`/`optional`, `variadic`, `repeatable`, `negated`, `defaultValue`) |
| `mnci kinds --json` | what `mnci add <kind>` accepts: `kind` (`node-app`, `go-lib`, `python-package`, …, one per language plugin and kind), `language`, `label`, `description`, `flags` |
| `mnci projects --json` | each project: `name`, `dir`, `kind`, `ecosystem` (`npm`, `pip`, `go`, or the plugin's name), `targets` (its tasks), and also `root`, `language`, `plugin`, `stack`, `deps`, `tasks` |
| `mnci info --json` | `cli` (`version`, `latest` — the newest release, or null when it cannot be found out quickly — and `updateAvailable`) and `workspace` (`root` and `mnci.toml` as JSON), or `"workspace": null` outside a workspace |

`mnci doctor --json` prints `findings` (`check`, `ok`, `detail`, `remedy`,
`warning`, `source`) with `passed`, `failed` and `warnings` counts.

`mnci add node-app web` is `mnci add app web --lang node`: the combined kinds
`mnci kinds` lists are accepted as they are.

Set `MNCI_NO_UPDATE_CHECK=1` to stop `mnci info` asking GitHub for the newest
release.
