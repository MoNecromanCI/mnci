package dependencies

import (
	"strings"
	"testing"
)

const pyproject = `[project]
name = "acme-api"
dependencies = []  # runtime

[tool.smith]
kind = "app"
`

func TestRequirementName(t *testing.T) {
	for spec, want := range map[string]string{
		"requests":                            "requests",
		"Requests_OAuthlib>=1.3":              "requests-oauthlib",
		"zope.interface ; python_version>'3'": "zope-interface",
		"pydantic[email]~=2.0":                "pydantic",
		"pkg @ https://example.com/p.whl":     "pkg",
	} {
		if got := RequirementName(spec); got != want {
			t.Errorf("%s: %s, want %s", spec, got, want)
		}
	}
}

func TestAddRequirements(t *testing.T) {
	got := AddRequirements(pyproject, "project", "dependencies", []string{"requests>=2.31", "Pydantic"})
	if !strings.Contains(got, `dependencies = ["requests>=2.31", "Pydantic"]  # runtime`) {
		t.Fatalf("single line:\n%s", got)
	}
	got = AddRequirements(got, "project", "dependencies", []string{"requests>=2.32"})
	if !strings.Contains(got, `["requests>=2.32", "Pydantic"]`) {
		t.Fatalf("a newer spec replaces the old one:\n%s", got)
	}
	got = AddRequirements(got, "project.optional-dependencies", "dev", []string{"pytest-cov"})
	if !strings.HasSuffix(got, "kind = \"app\"\n\n[project.optional-dependencies]\ndev = [\"pytest-cov\"]\n") {
		t.Fatalf("new table:\n%s", got)
	}
	multi := "[project]\ndependencies = [\n    \"httpx\",  # client\n]\n"
	if got := AddRequirements(multi, "project", "dependencies", []string{"rich"}); got != "[project]\ndependencies = [\n    \"httpx\",\n    \"rich\",\n]\n" {
		t.Fatalf("multi-line:\n%s", got)
	}
	if got := AddRequirements("[project]\nname = \"x\"\n", "project", "dependencies", []string{"rich"}); got != "[project]\ndependencies = [\"rich\"]\nname = \"x\"\n" {
		t.Fatalf("new key:\n%s", got)
	}
}

func TestRemoveRequirements(t *testing.T) {
	src := AddRequirements(pyproject, "project", "dependencies", []string{"requests>=2.31", "rich"})
	got, removed := RemoveRequirements(src, "project", "dependencies", []string{"Requests"})
	if strings.Join(removed, ",") != "requests" || !strings.Contains(got, `dependencies = ["rich"]  # runtime`) {
		t.Fatalf("%v\n%s", removed, got)
	}
	if _, removed := RemoveRequirements(src, "project", "dependencies", []string{"flask"}); removed != nil {
		t.Fatalf("removed %v", removed)
	}
}
