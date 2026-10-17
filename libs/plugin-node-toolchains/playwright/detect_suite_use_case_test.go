package playwright

import "testing"

func TestDetectFindsALibraryHoistedToTheWorkspaceRoot(t *testing.T) {
	ws := t.TempDir()
	// npm workspaces hoist shared dev tools: the suite declares nothing, the
	// root declares playwright.
	writeProject(t, ws, ".", `{"private": true, "workspaces": ["packages/*"], "devDependencies": {"playwright": "^1.63.0"}}`)
	writeProject(t, ws, "packages/office-reader", `{"name": "office-reader", "scripts": {"e2e": "jest -c jest.e2e.config.cjs"}}`)
	if got := Detect(ws, "packages/office-reader", true).Flavour; got != Library {
		t.Errorf("flavour %q, want library", got)
	}
	if got := Detect(ws, "packages/office-reader", false).Flavour; got != None {
		t.Errorf("without an e2e script a root playwright says nothing about this project: %q", got)
	}

	core := t.TempDir()
	writeProject(t, core, ".", `{"private": true, "devDependencies": {"playwright-core": "^1.63.0"}}`)
	writeProject(t, core, "apps/desktop", `{"name": "desktop"}`)
	if got := Detect(core, "apps/desktop", true).Flavour; got != Library {
		t.Errorf("playwright-core at the root counts too: %q", got)
	}

	none := t.TempDir()
	writeProject(t, none, ".", `{"private": true, "devDependencies": {"jest": "^30.0.0"}}`)
	writeProject(t, none, "apps/api", `{"name": "api"}`)
	if got := Detect(none, "apps/api", true).Flavour; got != None {
		t.Errorf("an e2e script with no Playwright anywhere is not a Playwright suite: %q", got)
	}
}
