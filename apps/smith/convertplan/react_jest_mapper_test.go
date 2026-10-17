package convertplan

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
)

const studioJest = `module.exports = {
  displayName:     '@opencraw/studio-ui',
  preset:          '../../jest.preset.js',
  testEnvironment: 'jsdom',
  transform:       {
    '^(?!.*\\.(js|jsx|ts|tsx|css|json)$)': '@nx/react/plugins/jest',
    '^.+\\.[tj]sx?$':                      ['babel-jest', { presets: ['@nx/react/babel'] }],
  },
}
`

func TestRewriteNxReactJestWritesOutTheTransformerAndPresets(t *testing.T) {
	out, ok := RewriteNxReactJest(studioJest, "apps/studio-ui")
	if !ok {
		t.Fatal("not rewritten")
	}
	if !strings.Contains(out, `: '<rootDir>/../../jest.asset-transform.cjs',`) {
		t.Errorf("transformer:\n%s", out)
	}
	if !strings.Contains(out, `presets: [['@babel/preset-env', { targets: { node: 'current' }, loose: true }], ['@babel/preset-typescript', { allowDeclareFields: true }], ['@babel/preset-react', { runtime: 'automatic', development: true }]]`) {
		t.Errorf("presets:\n%s", out)
	}
	if _, ok := RewriteNxReactJest(strings.Replace(studioJest, "testEnvironment", "resolver: '@nx/jest/plugins/resolver', testEnvironment", 1), "apps/studio-ui"); ok {
		t.Error("another Nx import is left to the user")
	}
}

func TestPlanRewritesAReactJestConfig(t *testing.T) {
	ui := project("@acme/ui", "apps/ui", "app", []string{"package.json", "jest.config.cts"}, nil)
	w := &nxworkspace.Workspace{Projects: []*nxworkspace.Project{ui}}
	plan := MakePlan(w, Input{DirName: "acme", Files: []string{"apps/ui/jest.config.cts"}, Read: func(f string) string {
		if f == "apps/ui/jest.config.cts" {
			return studioJest
		}
		return ""
	}})
	if !strings.Contains(plan.Write["apps/ui/jest.config.cts"], "jest.asset-transform.cjs") || !strings.Contains(plan.Write[AssetTransformFile], "ReactComponent") {
		t.Fatalf("writes %v", plan.WriteFiles)
	}
	if _, ok := plan.AddDevDependencies["@babel/preset-react"]; !ok {
		t.Fatalf("dev dependencies %v", plan.AddDevDependencies)
	}
	if manual := findings(plan, Manual); strings.Contains(manual, "jest.config") {
		t.Fatalf("manual:\n%s", manual)
	}
}
