package scaffold

const configTemplate = `# mnci workspace configuration. Every setting below has a default; most
# workspaces need nothing more than a name. See docs/configuration.md.
schema = 1

[workspace]
name = %q
# npm scope for Node packages
scope = %q
# Go module path prefix for Go projects
module_prefix = %q

[ci]
# Pipelines to generate: "github", "azure" or both.
providers = [%s]

# Conventions add rules on top of the language plugins. vertical-slices gives
# Go and Python projects a slice-check task, run before lint: role-named
# files, no junk-drawer folders. Node projects get the same rules from
# @mnci/eslint-config. See docs/conventions.md.
[conventions.vertical-slices]

# [release]
# branch = "main"     # releases run only from this branch
# remote = "origin"   # release tags are pushed here

# [cache]
# dir = ".mnci/cache"
# max_size_mb = 2048

# Override any task a plugin infers, for every project or for one:
# [tasks.test]
# cache = false
# [projects.web.tasks.build]
# depends_on = ["^build", "codegen"]
`

const gitignore = `# mnci
.mnci/

# build output
dist/
coverage/

# Node
node_modules/

# Python
.venv/
__pycache__/
*.egg-info/
.pytest_cache/
.ruff_cache/

# editors and OS
.DS_Store
.idea/
`

const gitattributes = `* text=auto eol=lf
*.png binary
*.jpg binary
*.zip binary
`

const readme = `# %s

A monorepo managed by [mnci](https://github.com/MoNecromanCI/mnci).

## Everyday commands

` + "```sh" + `
mnci add app web --lang node        # an app: deployed, never published
mnci add lib core --lang go         # an internal library: imported, never published
mnci add package sdk --lang python  # a package: published on release

mnci build | test | lint            # every project, in dependency order, cached
mnci test --affected                # only what changed since main
mnci graph                          # how projects depend on each other
mnci release --dry-run              # what the next release would publish
mnci doctor                         # check the workspace is healthy
` + "```" + `

Commits follow [Conventional Commits](https://www.conventionalcommits.org): ` + "`feat:`" + ` releases a
minor version, ` + "`fix:`" + ` a patch, and a breaking change a major one.
`
