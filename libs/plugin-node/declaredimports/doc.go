// Package declaredimports checks that a Node package declares what it
// imports. A published package's imports are resolved from whatever installs
// it, so one its package.json does not declare works in the workspace (npm
// hoists everything) and breaks for whoever installs it from npm. It replaces
// Nx's @nx/dependency-checks, which smith convert removes.
package declaredimports
