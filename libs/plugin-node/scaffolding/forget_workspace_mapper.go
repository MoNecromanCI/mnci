package scaffolding

import "github.com/MoNecromanCI/mnci/packages/plugin-sdk/jsondoc"

// ForgetWorkspace drops a removed workspace package from package-lock.json:
// its own entry and the node_modules link to it. npm install alone keeps it,
// marked extraneous, for as long as the lockfile lists it.
func ForgetWorkspace(lock []byte, root string) ([]byte, bool, error) {
	doc, err := jsondoc.ParseDocument(lock)
	if err != nil {
		return nil, false, err
	}
	packages := doc.Object("packages", false)
	if packages == nil {
		return lock, false, nil
	}
	changed := false
	for _, key := range packages.Keys() {
		entry := packages.Object(key, false)
		linked := false
		if entry != nil {
			resolved, _ := entry.String("resolved")
			linked = resolved == root
		}
		if key == root || linked {
			packages.Delete(key)
			changed = true
		}
	}
	if !changed {
		return lock, false, nil
	}
	return doc.Bytes(), true, nil
}
