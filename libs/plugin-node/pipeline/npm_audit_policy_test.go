package pipeline

import (
	"encoding/json"
	"reflect"
	"testing"
)

func advisory(name, severity, fix string) Advisory {
	return Advisory{Name: name, Severity: severity, FixAvailable: json.RawMessage(fix)}
}

func TestClassifyAuditBlocksOnlyAFixAPlainUpdateApplies(t *testing.T) {
	for _, c := range []struct {
		label  string
		found  Advisory
		blocks bool
	}{
		{"a moderate advisory with a plain fix", advisory("a", "moderate", "true"), true},
		{"a high advisory with a plain fix", advisory("a", "high", `{"isSemVerMajor":false}`), true},
		{"a critical advisory with a plain fix", advisory("a", "critical", "true"), true},
		{"a low advisory with a fix", advisory("a", "low", "true"), false},
		{"a high advisory with no fix upstream", advisory("a", "high", "false"), false},
		{"a high advisory whose only fix is a major bump", advisory("a", "high", `{"isSemVerMajor":true}`), false},
	} {
		v := ClassifyAudit(AuditReport{Vulnerabilities: map[string]Advisory{"a": c.found}})
		if got := len(v.Blocking) == 1; got != c.blocks {
			t.Errorf("%s: blocks = %v", c.label, got)
		}
	}
}

func TestClassifyAuditSaysWhyEachOtherAdvisoryDoesNotBlock(t *testing.T) {
	v := ClassifyAudit(AuditReport{Vulnerabilities: map[string]Advisory{
		"x": advisory("x", "high", "false"),
		"y": advisory("y", "high", `{"isSemVerMajor":true}`),
		"z": advisory("z", "low", "true"),
	}})
	want := []string{
		"  note [high] x - NO fix available upstream, nothing to do here",
		"  note [high] y - only a semver-major change would remove it (often a parent downgrade), not applied automatically",
		"  note [low] z - fix available, below the blocking threshold",
	}
	if !reflect.DeepEqual(v.Notes, want) {
		t.Errorf("notes = %q", v.Notes)
	}
}

func TestClassifyAuditHasNothingToSayWithoutAdvisories(t *testing.T) {
	if v := ClassifyAudit(AuditReport{}); v.Total != 0 || len(v.Blocking) != 0 || len(v.Notes) != 0 {
		t.Errorf("verdict = %+v", v)
	}
}

func TestPlanIsEmpty(t *testing.T) {
	for _, c := range []struct {
		label string
		plan  string
		empty bool
	}{
		{"a readable plan with nothing to add, change or remove", `{"add":[],"change":[],"remove":[],"audit":{}}`, true},
		{"a plan that adds something", `{"add":[{"name":"x"}],"change":[],"remove":[]}`, false},
		{"a plan that changes something", `{"add":[],"change":[{"name":"x"}],"remove":[]}`, false},
		{"a plan that moves a version", `{"add":[],"change":[{"from":{"name":"x","version":"1.0.0"},"to":{"name":"x","version":"1.0.1"}}],"remove":[]}`, false},
		// --omit=dev with node_modules installed plans to remove the dev packages.
		{"a plan that only removes packages", `{"add":[],"change":[],"remove":[{"name":"x"}]}`, true},
		{"output without the plan lists", `{"vulnerabilities":{}}`, false},
		{"a plan whose lists are not lists", `{"add":0,"change":0,"remove":0}`, false},
		// npm lists a workspace link as a change to the same version in a fresh checkout.
		{"a change that keeps its version", `{"add":[],"change":[{"from":{"name":"@acme/cli","version":"1.0.0","path":"/a"},"to":{"name":"@acme/cli","version":"1.0.0","path":"/b"}}],"remove":[]}`, true},
		// npm 11 prints the plan as text lines ahead of the JSON document.
		{"text ahead of an empty plan", "npm warn something\n{\n  \"add\": [],\n  \"change\": [],\n  \"remove\": []\n}", true},
		{"text ahead of a plan with a change", "change x 1.0.0 => 1.0.1\n{\n  \"add\": [],\n  \"change\": [{\"name\": \"x\"}],\n  \"remove\": []\n}", false},
		{"output that is not JSON", "npm error network", false},
		{"no output", "", false},
	} {
		if got := PlanIsEmpty(c.plan); got != c.empty {
			t.Errorf("%s: empty = %v", c.label, got)
		}
	}
}
