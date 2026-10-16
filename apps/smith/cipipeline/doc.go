// Package cipipeline generates the workspace's CI pipelines and runs their
// phases. Pipelines are derived from what is in the workspace: each plugin
// contributes its toolchain, setup, audit and credentials, mnci builds one
// provider-neutral model from them, and the GitHub Actions and Azure Pipelines
// renderers turn it into thin YAML that calls `mnci ci <phase>` — so the same
// logic runs on a laptop and in CI. Team steps between mnci:slot markers survive
// regeneration.
package cipipeline
