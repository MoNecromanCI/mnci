package pipeline

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
)

// AuditReport is what `npm audit --json` reports, as far as the gate reads it.
type AuditReport struct {
	Vulnerabilities map[string]Advisory `json:"vulnerabilities"`
}

// Advisory is one vulnerable package in an npm audit report.
type Advisory struct {
	Name     string `json:"name"`
	Severity string `json:"severity"`
	// FixAvailable is false, true, or an object saying whether the fix is a
	// semver-major change.
	FixAvailable json.RawMessage `json:"fixAvailable"`
}

// fix says whether the advisory has a fix, and whether that fix is only a
// semver-major change (often a parent downgrade).
func (a Advisory) fix() (available, majorOnly bool) {
	var plain bool
	if json.Unmarshal(a.FixAvailable, &plain) == nil {
		return plain, false
	}
	var object struct {
		IsSemVerMajor bool `json:"isSemVerMajor"`
	}
	if json.Unmarshal(a.FixAvailable, &object) == nil {
		return true, object.IsSemVerMajor
	}
	return false, false
}

// blockingSeverities are the severities that block; low and info never do.
var blockingSeverities = map[string]bool{"critical": true, "high": true, "moderate": true}

// AuditVerdict is which advisories of a report block, and why the rest do not.
type AuditVerdict struct {
	Blocking []Advisory
	Total    int
	Notes    []string
}

// ClassifyAudit decides which advisories block: one at moderate or above
// with a fix a plain update applies, which someone can act on. The rest pass
// with a note: no fix exists upstream, only a semver-major change removes it
// (never applied automatically), or the severity is below the threshold. A
// gate that failed on advisories nobody can fix would stay red and be
// ignored; the fix to apply is a targeted overrides entry, never
// `npm audit fix --force`.
func ClassifyAudit(r AuditReport) AuditVerdict {
	names := make([]string, 0, len(r.Vulnerabilities))
	for name := range r.Vulnerabilities {
		names = append(names, name)
	}
	sort.Strings(names)
	v := AuditVerdict{Total: len(names)}
	for _, name := range names {
		a := r.Vulnerabilities[name]
		if a.Name == "" {
			a.Name = name
		}
		available, majorOnly := a.fix()
		switch {
		case available && !majorOnly && blockingSeverities[a.Severity]:
			v.Blocking = append(v.Blocking, a)
		case !available:
			v.Notes = append(v.Notes, note(a, "NO fix available upstream, nothing to do here"))
		case majorOnly:
			v.Notes = append(v.Notes, note(a, "only a semver-major change would remove it (often a parent downgrade), not applied automatically"))
		default:
			v.Notes = append(v.Notes, note(a, "fix available, below the blocking threshold"))
		}
	}
	return v
}

func note(a Advisory, why string) string {
	return fmt.Sprintf("  note [%s] %s - %s", a.Severity, a.Name, why)
}

// planStart is where the JSON document starts: npm 11 prints the plan as
// text lines (`change x 1 => 2`) ahead of it.
var planStart = regexp.MustCompile(`(?m)^\{`)

// PlanIsEmpty reports whether `npm audit fix --dry-run --json` has nothing to
// apply. npm reports a fix for an advisory even when no patched release
// exists anywhere (a chain whose leaf has no patched version), and
// `npm audit fix` then plans no change; blocking on that blocks on something
// nobody can do. Only an added package or a moved version is a fix:
//   - a change that keeps its version does not count: in a fresh `npm ci`
//     checkout npm lists the workspace's own links as `pkg 1.0.0 => 1.0.0`;
//   - removals do not count: with node_modules installed, as in CI after
//     setup, `--omit=dev` plans to remove every dev and extraneous package,
//     which fixes nothing (measured with npm 11.6: a fixable semver@7.5.1
//     plans one change and three removals, an unfixable tree only removals).
//
// Output that is not a readable plan is not an empty one, so the gate keeps
// blocking rather than passing on a result it could not read.
func PlanIsEmpty(stdout string) bool {
	if loc := planStart.FindStringIndex(stdout); loc != nil {
		stdout = stdout[loc[0]:]
	}
	var plan struct {
		Add    *[]json.RawMessage `json:"add"`
		Change *[]planChange      `json:"change"`
		Remove *[]json.RawMessage `json:"remove"`
	}
	if json.Unmarshal([]byte(stdout), &plan) != nil || plan.Add == nil || plan.Change == nil || plan.Remove == nil {
		return false
	}
	if len(*plan.Add) > 0 {
		return false
	}
	for _, c := range *plan.Change {
		if !c.noOp() {
			return false
		}
	}
	return true
}

type planChange struct {
	From *planPackage `json:"from"`
	To   *planPackage `json:"to"`
}

type planPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// noOp reports whether a change leaves the package at the version it had;
// anything not shaped like that is a real change.
func (c planChange) noOp() bool {
	return c.From != nil && c.To != nil && c.From.Version != "" && c.From.Version == c.To.Version && c.From.Name == c.To.Name
}
