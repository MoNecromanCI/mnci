package doctor

import (
	"sort"
	"time"

	"github.com/MoNecromanCI/mnci/apps/mnci/cipipeline"
	"github.com/MoNecromanCI/mnci/apps/mnci/release"
	"github.com/MoNecromanCI/mnci/apps/mnci/releaselineage"
	"github.com/MoNecromanCI/mnci/apps/mnci/session"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Report is `mnci doctor --json`.
type Report struct {
	Findings []Finding `json:"findings"`
	Passed   int       `json:"passed"`
	Failed   int       `json:"failed"`
	Warnings int       `json:"warnings"`
}

// Finding is one check; Source is "mnci" or the plugin's name.
type Finding struct {
	protocol.Finding
	Source string `json:"source"`
}

// Run collects every finding.
func Run(s *session.Session) Report {
	var findings []Finding
	add := func(source string, fs ...protocol.Finding) {
		for _, f := range fs {
			findings = append(findings, Finding{Finding: f, Source: source})
		}
	}
	add("mnci", CoreChecks(s)...)
	add("mnci", lineageChecks(s)...)
	for _, c := range s.Plugins.Clients {
		if !c.Info.Capabilities.Doctor {
			continue
		}
		var projects []protocol.Project
		for _, p := range s.Graph.Projects {
			if p.ServedBy(c.Info.Plugin.Name) {
				projects = append(projects, p.Project)
			}
		}
		var res protocol.DoctorResult
		if err := c.Call(protocol.MethodDoctor, protocol.DoctorParams{Projects: projects}, &res, time.Minute); err != nil {
			add(c.Info.Plugin.Name, protocol.Finding{Check: "plugin " + c.Info.Plugin.Name + " answers doctor", Detail: err.Error(), Remedy: "report this to the plugin's maintainers"})
			continue
		}
		add(c.Info.Plugin.Name, res.Findings...)
	}
	report := Report{Findings: findings}
	for _, f := range findings {
		switch {
		case f.OK:
			report.Passed++
		case f.Warning:
			report.Warnings++
		default:
			report.Failed++
		}
	}
	sort.SliceStable(report.Findings, func(i, j int) bool { return !report.Findings[i].OK && report.Findings[j].OK })
	return report
}

// lineageChecks report release tags a release would refuse to start from:
// under an old name, or behind the registry (this asks each registry).
func lineageChecks(s *session.Session) []protocol.Finding {
	failed := map[string]bool{}
	var out []protocol.Finding
	for _, f := range release.CheckLineage(s, false) {
		failed[f.Check] = true
		out = append(out, protocol.Finding{Check: f.Check, Detail: f.Message, Warning: !f.Blocking, Remedy: "see the detail; mnci release refuses to run until it is fixed"})
	}
	for _, check := range []string{releaselineage.CheckResolve, releaselineage.CheckRegistry} {
		if !failed[check] {
			out = append(out, protocol.Finding{Check: check, OK: true})
		}
	}
	return out
}

// CoreChecks are mnci's own invariants.
func CoreChecks(s *session.Session) []protocol.Finding {
	var out []protocol.Finding
	ignored, _ := s.Git.Run("check-ignore", "-q", ".mnci/cache/x")
	_, err := s.Git.Run("check-ignore", ".mnci/cache/x")
	_ = ignored
	out = append(out, protocol.Finding{
		Check: "the cache directory is ignored by git", OK: err == nil,
		Detail: ".mnci/ is not in .gitignore, so cache files could be committed", Remedy: "add `.mnci/` to .gitignore",
	})
	out = append(out, protocol.Finding{
		Check: "the repository has commits", OK: s.Git.HasCommits(),
		Detail: "releases and affected detection need history", Remedy: "commit the workspace",
	})
	out = append(out, protocol.Finding{
		Check: "the clone has full history", OK: !s.Git.IsShallow(), Warning: true,
		Detail: "a shallow clone cannot compute versions or affected projects", Remedy: "`git fetch --unshallow` (in CI: checkout with fetch-depth 0)",
	})
	if c, err := cipipeline.Collect(s); err == nil {
		drift := cipipeline.Drift(s.Root, s.Config, cipipeline.NewModel(c, s.Version, s.Config.Release.Branch))
		detail := ""
		if len(drift) > 0 {
			detail = drift[0]
		}
		out = append(out, protocol.Finding{
			Check: "CI pipelines match the workspace", OK: len(drift) == 0,
			Detail: detail, Remedy: "run `mnci ci generate` and commit the result",
		})
	}
	for _, p := range s.Graph.Projects {
		missing := []string{}
		for _, t := range []string{"lint", "test"} {
			if _, ok := p.Tasks[t]; !ok {
				missing = append(missing, t)
			}
		}
		if len(missing) > 0 {
			out = append(out, protocol.Finding{
				Check: p.Name + " can be verified", Warning: true,
				Detail: p.Name + " has no " + missing[0] + " task, so CI never checks it", Remedy: "add the task to the project (or in mnci.toml under [projects." + p.Name + ".tasks])",
			})
		}
	}
	return out
}
