// Package scaffold creates workspaces (mnci new) and projects (mnci add). A
// project is one of three kinds — an app (deployed, never published), a lib
// (internal, never published) or a package (published) — in any language a
// plugin provides; the plugin writes the files, mnci picks the directory and
// regenerates the pipelines.
package scaffold
