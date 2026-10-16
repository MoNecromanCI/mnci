// Package plugindocker is the docker plugin for mnci: a deploy plugin above
// the language plugins. Configured as a convention ([conventions.docker]),
// it gives every app with a Dockerfile an image task, writes a Dockerfile
// for a new app it knows how to ship, and pushes the image to
// CONTAINER_REGISTRY when the app releases. Its code lives in the slice
// packages below; plugin is the entry point.
package plugindocker
