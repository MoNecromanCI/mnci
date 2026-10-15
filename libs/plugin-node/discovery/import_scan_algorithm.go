package discovery

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	sourceExtensions = map[string]bool{".ts": true, ".tsx": true, ".mts": true, ".cts": true, ".js": true, ".jsx": true, ".mjs": true, ".cjs": true}
	skippedDirs      = map[string]bool{"node_modules": true, "dist": true, "coverage": true, "build": true}
	specifierPattern = regexp.MustCompile(`(?:\bfrom\s*|\bimport\s*\(?\s*|\brequire\s*\(\s*)['"]([^'"\n]+)['"]`)
)

// ScanImports returns the bare package names imported anywhere under dir,
// sorted and unique. Relative paths and node: built-ins are ignored.
func ScanImports(dir string) ([]string, error) {
	seen := map[string]bool{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != dir && (skippedDirs[d.Name()] || strings.HasPrefix(d.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !sourceExtensions[filepath.Ext(path)] || strings.HasSuffix(path, ".d.ts") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, name := range PackageNames(string(src)) {
			seen[name] = true
		}
		return nil
	})
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out, err
}

// PackageNames extracts bare package names from one source file's imports.
func PackageNames(src string) []string {
	var out []string
	for _, m := range specifierPattern.FindAllStringSubmatch(src, -1) {
		if name := packageName(m[1]); name != "" {
			out = append(out, name)
		}
	}
	return out
}

func packageName(specifier string) string {
	if specifier == "" || strings.HasPrefix(specifier, ".") || strings.HasPrefix(specifier, "/") || strings.Contains(specifier, ":") {
		return ""
	}
	parts := strings.Split(specifier, "/")
	if strings.HasPrefix(specifier, "@") {
		if len(parts) < 2 {
			return ""
		}
		return parts[0] + "/" + parts[1]
	}
	return parts[0]
}
