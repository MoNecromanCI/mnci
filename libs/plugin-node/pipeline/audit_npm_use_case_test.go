package pipeline

import (
	"strings"
	"testing"
)

const (
	clean      = `{"vulnerabilities":{}}`
	fixable    = `{"vulnerabilities":{"lodash":{"name":"lodash","severity":"high","fixAvailable":true}}}`
	emptyPlan  = `{"add":[],"change":[],"remove":[]}`
	changePlan = `{"add":[],"change":[{"from":{"name":"lodash","version":"4.17.20"},"to":{"name":"lodash","version":"4.17.21"}}],"remove":[]}`
)

// npmPrinting answers each npm command with the given output: the whole
// tree's audit, the production tree's, and the dry run of audit fix.
func npmPrinting(whole, production, dryRun string, ran *[]string) NpmRunner {
	return func(args ...string) string {
		line := strings.Join(args, " ")
		*ran = append(*ran, line)
		switch {
		case strings.HasPrefix(line, "audit fix"):
			return dryRun
		case strings.Contains(line, "--omit=dev"):
			return production
		}
		return whole
	}
}

func audit(whole, production, dryRun string) (passed bool, logged string, ran []string) {
	var lines []string
	passed = AuditNpm(npmPrinting(whole, production, dryRun, &ran), func(s string) { lines = append(lines, s) })
	return passed, strings.Join(lines, "\n"), ran
}

func TestAuditPassesACleanWorkspaceAndSaysSo(t *testing.T) {
	passed, logged, _ := audit(clean, clean, "")
	if !passed || !strings.Contains(logged, "none actionable at moderate or above in production dependencies") {
		t.Errorf("passed %v, logged %q", passed, logged)
	}
}

func TestAuditFailsOnAProductionAdvisoryWithAFixAndSaysHowToFixIt(t *testing.T) {
	passed, logged, _ := audit(fixable, fixable, changePlan)
	if passed || !strings.Contains(logged, "BLOCKING [high] lodash - fix available") || !strings.Contains(logged, "overrides") {
		t.Errorf("passed %v, logged %q", passed, logged)
	}
}

func TestAuditDoesNotBlockOnAnAdvisoryOnlyInDevDependencies(t *testing.T) {
	passed, logged, _ := audit(fixable, clean, changePlan)
	if !passed || !strings.Contains(logged, "lodash - reaches only dev dependencies") {
		t.Errorf("passed %v, logged %q", passed, logged)
	}
}

func TestAuditDoesNotBlockOnAReportThatIsNotJSON(t *testing.T) {
	passed, logged, _ := audit("npm error network", "npm error network", "")
	if !passed || !strings.Contains(logged, "not blocking on a broken audit") {
		t.Errorf("passed %v, logged %q", passed, logged)
	}
}

func TestAuditPassesWhenNpmReportsAFixItCannotApply(t *testing.T) {
	passed, logged, _ := audit(fixable, fixable, emptyPlan)
	if !passed || !strings.Contains(logged, "npm audit fix has nothing to apply") {
		t.Errorf("passed %v, logged %q", passed, logged)
	}
}

func TestAuditStillBlocksWhenTheDryRunCannotBeRead(t *testing.T) {
	if passed, logged, _ := audit(fixable, fixable, "npm error"); passed {
		t.Errorf("passed, logged %q", logged)
	}
}

func TestAuditRunsNoDryRunWhenNothingWouldBlock(t *testing.T) {
	_, _, ran := audit(clean, clean, "")
	for _, line := range ran {
		if strings.HasPrefix(line, "audit fix") {
			t.Errorf("ran %q", line)
		}
	}
}
