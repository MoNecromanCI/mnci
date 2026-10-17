# Editors

Editors and tools build on four commands whose `--json` output is a stable
contract. They keep the shapes the mnci VS Code extension reads, which was
written against the old TypeScript mnci.

| Command | Prints |
|---|---|
| `smith commands --json` | every command: `name`, `aliases`, `group` (`workspace`, `projects`, `dependencies`, `pipeline`, `inspect`; an editor offers everything outside `inspect`), `description`, `arguments` (`name`, `description`, `required`, `variadic`) and `options` (`name`, `flags`, `short`, `description`, `takesValue`: `none`/`required`/`optional`, `variadic`, `repeatable`, `negated`, `defaultValue`) |
| `smith kinds --json` | what `smith add <kind>` accepts: `kind` (`node-app`, `go-lib`, `python-package`, …, one per language plugin and kind), `language`, `label`, `description`, `flags` |
| `smith projects --json` | each project: `name`, `dir`, `kind`, `ecosystem` (`npm`, `pip`, `go`, or the plugin's name), `targets` (its tasks), and also `root`, `language`, `plugin`, `stack`, `deps`, `tasks` |
| `smith info --json` | `cli` (`version`, `latest` — the newest release, or null when it cannot be found out quickly — and `updateAvailable`) and `workspace` (`root` and `smith.toml` as JSON), or `"workspace": null` outside a workspace |

`smith doctor --json` prints `findings` (`check`, `ok`, `detail`, `remedy`,
`warning`, `source`) with `passed`, `failed` and `warnings` counts.

`smith schema` prints the JSON Schema of `smith.toml` (see configuration).

`smith add node-app web` is `smith add app web --lang node`: the combined kinds
`smith kinds` lists are accepted as they are.

Set `SMITH_NO_UPDATE_CHECK=1` to stop `smith info` asking GitHub for the newest
release.

## AI assistants (MCP)

`smith mcp` serves the workspace over the Model Context Protocol on stdin and
stdout. Register it as an MCP server whose command is `smith mcp`, started in
the workspace. For example, in a client's JSON configuration:

```json
{ "mcpServers": { "smith": { "command": "smith", "args": ["mcp"] } } }
```

The read-only tools (`smith_projects`, `smith_graph`, `smith_affected`,
`smith_kinds`, `smith_stacks`, `smith_commands`, `smith_info`, `smith_doctor`,
`smith_release_preview`, `smith_updates`) are marked so a client may run them
without asking. The others change the workspace: `smith_run`,
`smith_add_project`, `smith_remove_project`, `smith_install`, `smith_uninstall`,
`smith_sync` and `smith_pipeline`. Each tool runs the smith command it stands for
and returns its output headed by that command line (`$ smith install -w web
zod`), so you can see and repeat exactly what was done. Values become single
arguments, and one that looks like an option is refused.
