// Package convertplan decides how an Nx workspace, or one made by the old
// mnci, becomes an mnci workspace: projects, kinds, tasks, release settings
// and the files to edit or remove. It reports everything it cannot map and
// changes nothing on disk.
package convertplan
