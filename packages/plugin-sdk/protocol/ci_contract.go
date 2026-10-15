package protocol

// CIParams gives the plugin the projects it owns so it can describe what CI
// needs for them.
type CIParams struct {
	Projects []Project `json:"projects"`
}

// CIResult is the plugin's contribution to every generated pipeline.
type CIResult struct {
	// Toolchains are rendered as provider-native setup steps.
	Toolchains []Toolchain `json:"toolchains,omitempty"`
	// Setup steps run in `mnci ci setup`, in order, from the workspace root.
	Setup []Step `json:"setup,omitempty"`
	// Audit steps run in `mnci ci audit`.
	Audit []Step `json:"audit,omitempty"`
	// Credentials names the environment variables `mnci ci release` needs.
	Credentials []Credential `json:"credentials,omitempty"`
}

// Toolchain is a language runtime the pipeline must install.
type Toolchain struct {
	// Name is one of node, python, go.
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Step is one command a CI phase runs.
type Step struct {
	Name    string   `json:"name"`
	Command []string `json:"command"`
	// Blocking false makes a failure a warning instead of an error.
	Blocking bool `json:"blocking"`
}

// Credential is a secret the release phase needs.
type Credential struct {
	Env         string `json:"env"`
	Description string `json:"description"`
	// Required false means release works without it (e.g. a dry run target).
	Required bool `json:"required"`
}
