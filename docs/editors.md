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

`mnci schema` prints the JSON Schema of `mnci.toml` (see configuration).

`mnci add node-app web` is `mnci add app web --lang node`: the combined kinds
`mnci kinds` lists are accepted as they are.

Set `MNCI_NO_UPDATE_CHECK=1` to stop `mnci info` asking GitHub for the newest
release.

## AI assistants (MCP)

`mnci mcp` serves the workspace over the Model Context Protocol on stdin and
stdout. Register it as an MCP server whose command is `mnci mcp`, started in
the workspace. For example, in a client's JSON configuration:

```json
{ "mcpServers": { "mnci": { "command": "mnci", "args": ["mcp"] } } }
```

The read-only tools (`mnci_projects`, `mnci_graph`, `mnci_affected`,
`mnci_kinds`, `mnci_stacks`, `mnci_commands`, `mnci_info`, `mnci_doctor`,
`mnci_release_preview`, `mnci_updates`) are marked so a client may run them
without asking. The others change the workspace: `mnci_run`,
`mnci_add_project`, `mnci_remove_project`, `mnci_install`, `mnci_uninstall`,
`mnci_sync` and `mnci_pipeline`. Each tool runs the mnci command it stands for
and returns its output headed by that command line (`$ mnci install -w web
zod`), so you can see and repeat exactly what was done. Values become single
arguments, and one that looks like an option is refused.
