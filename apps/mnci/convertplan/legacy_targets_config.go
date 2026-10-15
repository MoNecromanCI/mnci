package convertplan

// legacyTargets are targets the old mnci wrote into every project of a kind,
// replaced by what mnci does itself.
var legacyTargets = map[string]string{
	"package":            "zipped the app into dist/drop for the pack phase; mnci has no pack phase",
	"package-all":        "zipped every platform's binary; mnci release builds an app's assets",
	"build-all":          "built every platform's binary; mnci release builds an app's assets",
	"nx-release-publish": "mnci release publishes",
}

// legacyRootScripts are root package.json scripts the old mnci wrote; they
// are removed rather than rewritten.
var legacyRootScripts = map[string]string{
	"python:install": "mnci ci setup creates the Python environment",
}

// legacyProjectScriptSuffixes are the per-project root scripts the old mnci
// wrote (<project>:<suffix>); `mnci run <task> -p <project>` replaces them.
var legacyProjectScriptSuffixes = []string{":qa", ":build", ":test", ":lint", ":lint:fix", ":start", ":build:dev", ":dev", ":typecheck"}

// legacyTools are files under tools/ the old mnci wrote for Nx release.
var legacyTools = []string{"tools/go-app-release.cjs", "tools/csharp-version-actions.cjs", "tools/cargo-version-actions.cjs", "tools/maven-version-actions.cjs", "tools/gradle-version-actions.cjs", "tools/composer-version-actions.cjs", "tools/ruby-version-actions.cjs", "tools/cmake-version-actions.cjs"}
