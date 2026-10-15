package scaffold

import (
	"fmt"
	"regexp"
)

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)

// ValidateName accepts kebab-case names that are valid as a directory, an npm
// package name, a Go path element and (with - → _) a Python import name.
func ValidateName(kind, name string) error {
	if len(name) > 50 || !namePattern.MatchString(name) {
		return fmt.Errorf("%s name %q must be kebab-case: lowercase letters and digits, words joined by single hyphens, starting with a letter (e.g. web-app)", kind, name)
	}
	return nil
}
