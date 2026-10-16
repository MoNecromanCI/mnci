package discovery

import (
	"regexp"
	"strings"
)

// GradleBuild is what mnci reads from a Gradle module: its build script
// (build.gradle.kts or build.gradle) and the gradle.properties beside it.
type GradleBuild struct {
	// Path is the Gradle project path, :apps:api for apps/api.
	Path                       string
	Group, ArtifactID, Version string
	// Kind and Stack are the mnci.kind and mnci.stack properties.
	Kind, Stack string
	// Projects are the paths of the workspace projects it depends on.
	Projects []string
	// Lint are the lint tasks its plugins add (ktlint, Checkstyle).
	Lint []string
	// BuildDir is the workspace-relative directory it builds to, when its
	// script moves the build to dist/.
	BuildDir string
}

// Coordinates is groupId:artifactId when the module publishes, else empty.
func (g GradleBuild) Coordinates() string {
	if g.Group == "" || g.ArtifactID == "" {
		return ""
	}
	return g.Group + ":" + g.ArtifactID
}

// ProjectPath is the Gradle path of the project in root: settings include
// it by that path, so its directory is the path's segments.
func ProjectPath(root string) string { return ":" + strings.ReplaceAll(root, "/", ":") }

var (
	assignment        = regexp.MustCompile(`(?m)^\s*(group|version|artifactId)\s*=\s*["']([^"']*)["']`)
	projectDependency = regexp.MustCompile(`project\(\s*(?:path\s*[=:]\s*)?["'](:[^"']*)["']`)
	gradleProperty    = regexp.MustCompile(`(?m)^\s*(mnci\.kind|mnci\.stack|group|version)\s*[=:]\s*(\S+)`)
	distBuildDir      = regexp.MustCompile(`["'](` + regexp.QuoteMeta("dist/gradle/") + `[^"']+)["']`)
)

// ParseGradleBuild reads a module's build script and gradle.properties; the
// script's group and version win over the properties'.
func ParseGradleBuild(root string, script, properties []byte) GradleBuild {
	b := GradleBuild{Path: ProjectPath(root)}
	for _, m := range gradleProperty.FindAllStringSubmatch(string(properties), -1) {
		switch m[1] {
		case "mnci.kind":
			b.Kind = m[2]
		case "mnci.stack":
			b.Stack = m[2]
		case "group":
			b.Group = m[2]
		case "version":
			b.Version = m[2]
		}
	}
	src := string(script)
	for _, m := range assignment.FindAllStringSubmatch(src, -1) {
		switch m[1] {
		case "group":
			b.Group = m[2]
		case "version":
			b.Version = m[2]
		case "artifactId":
			b.ArtifactID = m[2]
		}
	}
	for _, m := range projectDependency.FindAllStringSubmatch(src, -1) {
		b.Projects = append(b.Projects, m[1])
	}
	switch {
	case strings.Contains(src, "org.jlleitschuh.gradle.ktlint"):
		b.Lint = []string{"ktlintCheck"}
	case strings.Contains(src, "checkstyle"):
		b.Lint = []string{"checkstyleMain", "checkstyleTest"}
	}
	if m := distBuildDir.FindStringSubmatch(src); m != nil {
		b.BuildDir = m[1]
	}
	return b
}
