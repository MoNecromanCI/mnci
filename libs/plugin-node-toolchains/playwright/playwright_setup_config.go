package playwright

import (
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tooling"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// testVersion is the @playwright/test a new project starts with.
const testVersion = "^1.63.0"

// reportsToken is replaced by the project's ReportsDir, relative to the
// project, so reports and traces land in the e2e task's outputs.
const reportsToken = "[[.Reports]]"

const config = `import { defineConfig, devices } from '@playwright/test'

// The end-to-end suite. ` + "`smith run e2e`" + ` installs the browsers, then runs it.
// Reports and traces go to the workspace's dist folder for this project.
export default defineConfig({
  testDir:   'e2e',
  // *.e2e.ts, so the unit test runner (Vitest, Jest) never picks these up.
  testMatch: '**/*.e2e.ts',
  outputDir: '[[.Reports]]/test-results',
  reporter:  [['list'], ['html', { outputFolder: '[[.Reports]]/report', open: 'never' }]],
  retries:   process.env.CI ? 2 : 0,
  use:       {
    // Set E2E_BASE_URL to the address of the app under test.
    baseURL: process.env.E2E_BASE_URL,
    trace:   'on-first-retry',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
})
`

const smoke = `import { expect, test } from '@playwright/test'

// Replace with a test of the app: page.goto('/') once E2E_BASE_URL is set.
test('a browser renders a page', async ({ page }) => {
  await page.setContent('<h1>Hello, world!</h1>')
  await expect(page.getByRole('heading')).toHaveText('Hello, world!')
})
`

func setup(_ string, params protocol.GenerateParams) tooling.Setup {
	reports := ReportsDirFromProject(params.Directory)
	return tooling.Setup{
		DevDependencies: map[string]string{"@playwright/test": testVersion},
		Files: map[string]string{
			"playwright.config.ts": strings.ReplaceAll(config, reportsToken, reports),
			"e2e/smoke.e2e.ts":     smoke,
		},
		Scripts: map[string]string{TaskName: "playwright test"},
	}
}
