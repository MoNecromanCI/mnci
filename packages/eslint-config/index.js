import astro from './configs/astro.config.js'
import base from './configs/base.config.js'
import css from './configs/css.config.js'
import dependencyChecks from './configs/dependency-checks.config.js'
import houseStyle from './configs/house-style.config.js'
import html from './configs/html.config.js'
import importGraph from './configs/import-graph.config.js'
import jest from './configs/tests.config.js'
import json from './configs/json.config.js'
import markdown from './configs/markdown.config.js'
import mdx from './configs/mdx.config.js'
import react from './configs/react.config.js'
import regexp from './configs/regexp.config.js'
import standard from './configs/standard.config.js'
import svelte from './configs/svelte.config.js'
import toml from './configs/toml.config.js'
import typeAware from './configs/type-aware.config.js'
import typescript from './configs/typescript.config.js'
import browserAutomation from './configs/browser-automation.config.js'
import verticalSlices from './configs/vertical-slices.config.js'
import yaml from './configs/yaml.config.js'

export { default as astro } from './configs/astro.config.js'
export { default as base } from './configs/base.config.js'
export { default as css } from './configs/css.config.js'
export { default as dependencyChecks } from './configs/dependency-checks.config.js'
// The rule's options, separately: ESLint REPLACES rule options rather than
// merging them, so an override that does not spread these loses `ignoredFiles`
// and starts reporting `@nx/rollup` as missing from every publishable
// package's dependencies. See the remarks on the function.
export { dependencyChecksOptions } from './configs/dependency-checks.config.js'
export { default as houseStyle } from './configs/house-style.config.js'
export { default as html } from './configs/html.config.js'
export { default as importGraph } from './configs/import-graph.config.js'
export { default as jest } from './configs/tests.config.js'
export { default as json } from './configs/json.config.js'
export { default as markdown } from './configs/markdown.config.js'
export { default as mdx } from './configs/mdx.config.js'
export { default as react } from './configs/react.config.js'
export { default as regexp } from './configs/regexp.config.js'
export { default as standard } from './configs/standard.config.js'
export { default as svelte } from './configs/svelte.config.js'
export { default as toml } from './configs/toml.config.js'
export { default as typeAware } from './configs/type-aware.config.js'
export { default as typescript } from './configs/typescript.config.js'
export { default as browserAutomation } from './configs/browser-automation.config.js'
export { default as verticalSlices } from './configs/vertical-slices.config.js'
export { default as yaml } from './configs/yaml.config.js'

/** Paths never worth linting, in any generated workspace. */
export const ignores = [
  '**/dist',
  '**/dist-dev',
  '**/dist-uat',
  '**/dist-prod',
  '**/out-tsc',
  '**/coverage',
  '**/.nx',
  '**/node_modules',
  '**/build',
  '**/.next',
  // Build artifacts, and the omission that made the generated root `lint` target
  // fail on a fresh workspace. `@nx/esbuild` writes intermediates here — for a
  // node-app, `tmp/<scope>/<name>/main-with-require-overrides.js` — and the root
  // lint target has no project directory to hide behind, so it linted them: 39
  // errors on generated code the user never wrote. `.prettierignore` has always
  // listed `tmp`, which is exactly why nothing noticed the asymmetry.
  '**/tmp',
  '**/vite.config.*.timestamp*',
  // Non-JS ecosystems smith supports; each has its own linter (ruff,
  // golangci-lint, flutter analyze) wired as that project's `lint` target.
  '**/.venv',
  '**/__pycache__',
  '**/.dart_tool',
  // Composer's and Bundler's install folders. A PHP project's `vendor/` holds every dependency's own README and
  // frontend scripts (a Laravel app's alone is 23,000 findings), and nothing else ignores it.
  '**/vendor',
  // A Rails app's PWA templates (`app/views/pwa/*.js`): comment-only files Rails renders as views and serves as classic
  // scripts. `unicorn/no-empty-file` rejects them, and `export {}` would turn a service worker into a module.
  '**/app/views/pwa',
  // A VS Code extension's integration tests (#244): `@vscode/test-cli` downloads a whole VS Code
  // into `.vscode-test/` (thousands of bundled JS files), and `tsc` writes the compiled tests to
  // `out-integration/`. Linting them hung `smith upgrade`'s `eslint --fix` for minutes in a 6 GB
  // process, found when the e2e first ran an extension's integration tests for real.
  '**/.vscode-test',
  '**/out-integration',
]

/**
 * The complete smith ESLint config: one root config for the whole workspace.
 *
 * @remarks
 * Composition order matters. `standard` is applied LAST because it IS the
 * formatting opinion, and nothing may follow that disables it. Where
 * `eslint-config-prettier` used to sit — switching every stylistic rule off so
 * a formatter could own them — there is now nothing, because there is no
 * formatter. Re-introducing it would silently disable all sixty Standard rules
 * and leave the workspace with no style enforcement at all: a disabled rule
 * reports nothing, so `lint` would pass
 * mutually unsatisfiable. See `configs/stylistic.js`.
 *
 * @param options - Composition options. Pass `workspaceRoot` (normally
 * `import.meta.dirname` from the root config) to enable the
 * `@nx/dependency-checks` block, which scans for `private: true` manifests.
 * Omit it in a workspace with no publishable npm packages. Pass
 * `verticalSlices: true` - the globs of the projects that follow it, or
 * `{ files, roles }` to extend the role vocabulary - to
 * change which projects the vertical-slice rules apply to (see `configs/vertical-slices.config.js`).
 * They are ON by default - smith's architecture is vertical feature slices - and
 * `verticalSlices: false` opts out. Pass
 * `browserAutomation: true` - or globs - in a Playwright or Puppeteer project,
 * so an in-page callback like `page.evaluate(() => document.title)` can see
 * browser globals; `unicorn/isolated-functions` reports every one of them
 * otherwise, and it is right to (see `configs/browser-automation.config.js`).
 * @returns The flat config array.
 */
export default function monosmith (options = {}) {
  const { workspaceRoot, verticalSlices: slices = true, browserAutomation: browser } = options

  return [
    { name: 'monosmith/ignores', ignores },
    ...base,
    ...typescript,
    ...typeAware,
    ...importGraph,
    ...react,
    ...regexp,
    ...json,
    ...yaml,
    ...toml,
    ...markdown,
    ...css,
    ...html,
    ...svelte,
    ...astro,
    ...mdx,
    ...jest,
    ...(workspaceRoot ? dependencyChecks(workspaceRoot) : []),
    ...(slices ? verticalSlices(slices === true ? undefined : slices) : []),
    // After the blocks that set Node's globals, so these ADD to them rather
    // than being overwritten by them.
    ...(browser ? browserAutomation(browser === true ? undefined : browser) : []),
    // LAST, and nothing may follow that disables it: this block IS the
    // formatting opinion now. `eslint-config-prettier` used to sit here to
    // switch every stylistic rule off for a formatter to own; with no
    // formatter, composing it would silently disable all 62 Standard rules.
    ...standard,
    // After `standard`, and only ever after it: these are the deliberate
    // departures from Standard, so they have to win. Composed as their own
    // named block rather than edited into the port above, which stays a
    // faithful extraction from neostandard.
    ...houseStyle,
  ]
}
