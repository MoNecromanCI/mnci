package layout

// Roles are the role suffixes a file may carry.
var Roles = []string{"handler", "use_case", "algorithm", "policy", "model", "contract", "mapper", "validator", "repository", "client", "store", "middleware", "error", "config", "enum"}

// ForbiddenFolders are junk drawers: a folder is named for the outcome it
// serves.
var ForbiddenFolders = map[string]bool{"util": true, "utils": true, "helper": true, "helpers": true, "common": true, "shared": true, "manager": true, "processor": true, "data": true, "misc": true}

// skippedFolders hold no hand-authored code, or only test data.
var skippedFolders = map[string]bool{"vendor": true, "testdata": true, "fixtures": true, "node_modules": true, "dist": true, "build": true, "__pycache__": true}

// SliceSize is the number of production files at which a slice is reviewed:
// it usually holds two outcomes and should become two slices.
const SliceSize = 12

// goPlatforms are the build-constraint suffixes a Go file name may end in.
var goPlatforms = []string{"windows", "linux", "darwin", "freebsd", "js", "wasm", "amd64", "arm64", "386", "arm"}
