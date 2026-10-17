// Package playwright is the playwright toolchain plugin: it owns a Node
// project's e2e task, installs the browsers that task needs before it runs,
// runs it under a virtual display on Linux when the project asks for one,
// and starts a new project with a Playwright Test suite.
package playwright
