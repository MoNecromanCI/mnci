package layout

// skippedFolders hold no hand-authored code, or only test data, in any
// language.
var skippedFolders = map[string]bool{"vendor": true, "testdata": true, "fixtures": true, "node_modules": true, "dist": true, "build": true, "__pycache__": true}

// SliceSize is the number of production files at which a slice is reviewed:
// it usually holds two outcomes and should become two slices.
const SliceSize = 12
