// Package scaffolding generates Node apps, libs and packages, laid out as
// vertical slices, from a template set: the node plugin's own plain
// JavaScript one, or another plugin's (typescript). The first Node project
// also brings the workspace's Node tooling: the root package.json with npm
// workspaces and its devDependencies, and the template set's root files.
package scaffolding
