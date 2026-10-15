# Decisions so far

Product north star: a super simple monorepo tool where you ask for a new app, an internal package or a
publishable package and it just works; CI and release are first-class and everything is generated.

| # | Decision |
|---|---|
| 1 | From scratch in Go; Nx and Please are idea sources, not forks. Current TypeScript mnci scaffolds and bootstraps this repo. |
| 2 | Language-agnostic core: discovery, task graph, hashing, cache, scheduler, affected, release, TUI, doctor. |
| 3 | Plugins are out-of-process, JSON-RPC 2.0 over stdio, versioned; kinds: language plugins and convention plugins; all live in this monorepo as separate packages. |
| 4 | Cache is content-addressed static files behind a `Storage` interface; local directory first, remote backends later. |
| 5 | Native Windows is first-class. |
| 6 | v1 languages: Node, Go, Python. Others are post-v1 waves. |
| 7 | Generators are plugin-owned: templates plus ecosystem scaffolders (`flutter create`, `dotnet new`, …). |
| 8 | Configuration is TOML, deliberately minimal, with provenance (`explain`) and TUI editing that preserves comments. Fallback if round-trip editing fails the spike: JSONC + schema. |
| 9 | Pipelines are generated from the workspace (plugins contribute setup/audit/verify/pack/release steps) through a provider-neutral model rendered to Azure Pipelines and GitHub Actions. |
| 10 | Release is resumable, idempotent, dependency-ordered, previewable (`--dry-run` is the same code path). |
| 11 | Distribution (npm, PyPI, Homebrew, Scoop/winget) wraps one static binary; v1 must not block it. |
| 12 | If current mnci cannot scaffold something needed in this monorepo, create it by hand and file a `gap-in-current-mnci` issue. |
