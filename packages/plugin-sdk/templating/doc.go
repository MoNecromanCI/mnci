// Package templating renders a set of file templates into a new project. It
// uses [[ ]] delimiters so templates can contain GitHub Actions, Jinja or Go
// template syntax verbatim, and it never overwrites an existing file.
package templating
