package release

import (
	"strings"
	"testing"
)

func TestMarkdownNamesTheMirror(t *testing.T) {
	result := ReportResult{Item: Item{Project: "client", Next: "1.2.0"}, Status: "released", Mirror: "acme/client-php@v1.2.0"}
	if got := (Report{Results: []ReportResult{result}}).Markdown(); !strings.Contains(got, "Mirrored client to acme/client-php@v1.2.0") {
		t.Errorf("release:\n%s", got)
	}
	if got := (Report{DryRun: true, Results: []ReportResult{result}}).Markdown(); !strings.Contains(got, "Would mirror client to acme/client-php@v1.2.0") {
		t.Errorf("preview:\n%s", got)
	}
	if got := (Report{Results: []ReportResult{{Item: Item{Project: "web"}}}}).Markdown(); strings.Contains(got, "irror") {
		t.Errorf("no mirror, no line:\n%s", got)
	}
}
