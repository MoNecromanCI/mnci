package declaredimports

import "fmt"

// UndeclaredImport is one package a project's file imports without
// declaring it where it must be.
type UndeclaredImport struct {
	// File is the importing file, relative to the project.
	File string
	// Package is the imported package's name.
	Package string
	// DevOnly is true when the package is declared, but only in
	// devDependencies, which installing the package does not install.
	DevOnly bool
}

func (u UndeclaredImport) String() string {
	if u.DevOnly {
		return fmt.Sprintf("%s imports %s, declared only in devDependencies", u.File, u.Package)
	}
	return fmt.Sprintf("%s imports %s, which package.json does not declare", u.File, u.Package)
}
