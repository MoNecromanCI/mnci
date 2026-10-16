package protocol

// DoctorParams gives the plugin its projects to check.
type DoctorParams struct {
	Projects []Project `json:"projects"`
}

// DoctorResult carries the plugin's findings.
type DoctorResult struct {
	Findings []Finding `json:"findings"`
}

// Finding is one check. Every failed finding names its remedy.
type Finding struct {
	Check  string `json:"check"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
	Remedy string `json:"remedy,omitempty"`
	// Warning true reports a problem without failing doctor.
	Warning bool `json:"warning,omitempty"`
	// Fix is the change that clears a failed finding, when the plugin can
	// write it; smith doctor --fix shows it as a diff before applying it.
	Fix *Fix `json:"fix,omitempty"`
}

// Fix is a change to files, as their whole new contents, so the host can show
// exactly what changes and apply it in any language.
type Fix struct {
	Description string `json:"description"`
	// Safe marks a fix that only restores what smith or the plugin owns and
	// loses nothing of the user's; doctor --fix applies only safe fixes.
	Safe  bool       `json:"safe"`
	Files []FileEdit `json:"files"`
}

// FileEdit is one file's new content, workspace-relative; Delete removes it.
type FileEdit struct {
	Path    string `json:"path"`
	Content string `json:"content,omitempty"`
	Delete  bool   `json:"delete,omitempty"`
}
