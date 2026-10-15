package release

import (
	"fmt"
	"strings"
	"time"
)

// Report is the machine-readable record of a release run, written to
// .mnci/release-report.json and printed by --json.
type Report struct {
	DryRun   bool           `json:"dryRun"`
	Started  time.Time      `json:"started"`
	Finished time.Time      `json:"finished"`
	Results  []ReportResult `json:"results"`
	// Unchanged lists packages with nothing to release.
	Unchanged []string `json:"unchanged,omitempty"`
	// Error is set when the release stopped early.
	Error string `json:"error,omitempty"`
}

// ReportResult is one package's outcome.
type ReportResult struct {
	Item
	// Status is planned (dry run), released, already-released or failed.
	Status    string   `json:"status"`
	Files     []string `json:"files,omitempty"`
	Published bool     `json:"published"`
	Publish   string   `json:"publish,omitempty"`
	Registry  string   `json:"registry,omitempty"`
	Pushed    bool     `json:"pushed"`
	// Assets are the files built for an app's release.
	Assets []string `json:"assets,omitempty"`
	// GitHubRelease is the URL of the tag's GitHub Release.
	GitHubRelease string `json:"githubRelease,omitempty"`
	Error         string `json:"error,omitempty"`
}

// Markdown renders the report for a CI step summary or a PR comment.
func (r Report) Markdown() string {
	var b strings.Builder
	title := "Release"
	if r.DryRun {
		title = "Release preview"
	}
	fmt.Fprintf(&b, "### %s\n\n", title)
	if len(r.Results) == 0 {
		b.WriteString("Nothing to release.\n")
	} else {
		b.WriteString("| Package | Version | Why | Status |\n|---|---|---|---|\n")
		for _, res := range r.Results {
			from := res.Current
			if from == "" {
				from = "new"
			}
			fmt.Fprintf(&b, "| %s | %s → **%s** | %s | %s |\n", res.Project, from, res.Next, res.Reason, res.Status)
		}
	}
	if len(r.Unchanged) > 0 {
		fmt.Fprintf(&b, "\nUnchanged: %s\n", strings.Join(r.Unchanged, ", "))
	}
	if r.Error != "" {
		fmt.Fprintf(&b, "\n**Stopped:** %s\n\nRe-run the release: finished packages are kept and the rest resume.\n", r.Error)
	}
	return b.String()
}
