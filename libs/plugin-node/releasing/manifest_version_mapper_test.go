package releasing

import "testing"

func TestSetVersionsPreservesFormatting(t *testing.T) {
	src := "{\r\n    \"name\": \"@acme/sdk\",\r\n    \"version\": \"0.0.0\",\r\n    \"nested\": { \"version\": \"keep\" },\r\n    \"dependencies\": {\"@acme/core\": \"*\", \"zod\": \"^3\"},\r\n    \"peerDependencies\": {\"@acme/core\": \">=1\"},\r\n    \"list\": [\"version\", {\"version\": \"x\"}]\r\n}\r\n"
	got, err := SetVersions([]byte(src), "1.2.3", map[string]string{"@acme/core": "2.0.0"})
	if err != nil {
		t.Fatal(err)
	}
	want := "{\r\n    \"name\": \"@acme/sdk\",\r\n    \"version\": \"1.2.3\",\r\n    \"nested\": { \"version\": \"keep\" },\r\n    \"dependencies\": {\"@acme/core\": \"^2.0.0\", \"zod\": \"^3\"},\r\n    \"peerDependencies\": {\"@acme/core\": \">=1\"},\r\n    \"list\": [\"version\", {\"version\": \"x\"}]\r\n}\r\n"
	if string(got) != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
}

func TestSetVersionsRequiresVersion(t *testing.T) {
	if _, err := SetVersions([]byte(`{"name":"x"}`), "1.0.0", nil); err == nil {
		t.Fatal("expected an error")
	}
}

func TestSetVersionsHandlesEscapes(t *testing.T) {
	got, err := SetVersions([]byte(`{"description":"a \"quoted\" {brace}","version":"0.1.0"}`), "0.2.0", nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"description":"a \"quoted\" {brace}","version":"0.2.0"}` {
		t.Fatalf("got %s", got)
	}
}
