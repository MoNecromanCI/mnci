package releasing

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func fakeNpm(t *testing.T, published string) *[]string {
	t.Helper()
	var calls []string
	npm.output = func(_ string, argv []string) (string, error) {
		calls = append(calls, strings.Join(argv, " "))
		if argv[1] == "view" {
			return published, nil
		}
		return "https://registry.example/\n", nil
	}
	npm.run = func(_ string, argv []string) error { calls = append(calls, strings.Join(argv, " ")); return nil }
	return &calls
}

func TestPublishSkipsExistingVersion(t *testing.T) {
	calls := fakeNpm(t, "1.0.0\n")
	res, err := Publish("/ws", protocol.PublishParams{Project: protocol.Project{Root: "packages/sdk", Exports: []string{"@acme/sdk"}}, Version: "1.0.0"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Published || res.Skipped != "already published" || res.Registry != "https://registry.example/" {
		t.Fatalf("result %+v", res)
	}
	for _, c := range *calls {
		if strings.Contains(c, "publish") {
			t.Fatal("must not publish an existing version")
		}
	}
}

func TestPublishNewVersion(t *testing.T) {
	calls := fakeNpm(t, "")
	res, err := Publish("/ws", protocol.PublishParams{Project: protocol.Project{Root: "packages/sdk", Exports: []string{"@acme/sdk"}}, Version: "1.1.0"})
	if err != nil || !res.Published {
		t.Fatalf("res %+v err %v", res, err)
	}
	if last := (*calls)[len(*calls)-1]; last != "npm publish --access public" {
		t.Fatalf("last call %q", last)
	}
}

func TestPublishDryRunTouchesNothing(t *testing.T) {
	calls := fakeNpm(t, "")
	res, err := Publish("/ws", protocol.PublishParams{Project: protocol.Project{Root: "p", Exports: []string{"x"}}, Version: "1.0.0", DryRun: true})
	if err != nil || res.Skipped != "dry run" || len(*calls) != 1 {
		t.Fatalf("res %+v calls %v err %v", res, *calls, err)
	}
}
