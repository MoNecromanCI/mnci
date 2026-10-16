package discovery

// PackageManifest is the subset of package.json the plugin reads.
type PackageManifest struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	Private              bool              `json:"private"`
	Scripts              map[string]string `json:"scripts"`
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	PeerDependencies     map[string]string `json:"peerDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
	Engines              struct {
		VSCode string `json:"vscode"`
	} `json:"engines"`
	Mnci struct {
		Kind  string `json:"kind"`
		Stack string `json:"stack"`
	} `json:"mnci"`
}
