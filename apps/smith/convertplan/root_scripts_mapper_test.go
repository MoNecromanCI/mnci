package convertplan

import "testing"

func TestRewriteScript(t *testing.T) {
	names := map[string]string{"@acme/web": "web", "api": "api"}
	for in, want := range map[string]string{
		"nx run-many -t build":                                  "mnci run build",
		"nx run-many -t lint,test --projects=@acme/web,api":     "mnci run lint test -p web,api",
		"nx run-many --target=test --parallel=3":                "mnci run test",
		"npx nx affected -t typecheck,test,build":               "mnci run typecheck test build --affected",
		"mnci lint --affected && nx affected -t typecheck,test": "mnci run lint --affected && mnci run typecheck test --affected",
		"nx run api:serve":                                      "mnci run serve -p api",
		"nx run @acme/web:build:production":                     "mnci run build -p web",
		"nx build api":                                          "mnci run build -p api",
		"nx graph":                                              "mnci graph",
		"nx release --dry-run":                                  "mnci release --dry-run",
		"husky":                                                 "husky",
		"eslint . --fix --cache":                                "eslint . --fix --cache",
	} {
		got, ok := RewriteScript(in, names)
		if !ok || got != want {
			t.Errorf("%q: got %q (%v), want %q", in, got, ok, want)
		}
	}
	for _, in := range []string{"nx migrate latest", "nx run unknown:build", "nx release publish", "nx run-many -t build --exclude=x"} {
		if got, ok := RewriteScript(in, names); ok {
			t.Errorf("%q must not convert, got %q", in, got)
		}
	}
}
