// Package adoption is `mnci adopt`: it reads a repository that is not yet an
// mnci workspace (plain npm workspaces, Go modules, Python projects; an Nx
// workspace is mnci convert's) and reports what bringing it under mnci
// involves, each blocker and warning with the step that clears it. Two steps
// it can take itself: --tags creates the baseline release tags the lineage
// checks ask for, and --init writes mnci.toml and the pipelines.
package adoption
