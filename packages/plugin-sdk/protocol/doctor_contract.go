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
}
