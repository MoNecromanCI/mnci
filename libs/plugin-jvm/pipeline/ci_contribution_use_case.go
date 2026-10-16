package pipeline

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/toolrun"
)

// Versions installed when the workspace does not pin them; bump them
// deliberately.
const (
	JavaVersion   = "21"
	MavenVersion  = "3.9.11"
	GradleVersion = "9.8.1"
)

var release = regexp.MustCompile(`<maven\.compiler\.release>\s*(\d+)\s*</maven\.compiler\.release>`)

// Version is the Java release .java-version or the root pom's
// maven.compiler.release pins, else JavaVersion.
func Version(workspaceRoot string) string {
	if raw, err := os.ReadFile(filepath.Join(workspaceRoot, ".java-version")); err == nil {
		if v, _, _ := strings.Cut(strings.TrimSpace(string(raw)), "."); v != "" {
			return v
		}
	}
	if raw, err := os.ReadFile(filepath.Join(workspaceRoot, "pom.xml")); err == nil {
		if m := release.FindSubmatch(raw); m != nil {
			return string(m[1])
		}
	}
	return JavaVersion
}

// buildTools says which build tools the projects use.
type buildTools struct{ maven, gradle bool }

func usedBuildTools(workspaceRoot string, projects []protocol.Project) buildTools {
	var used buildTools
	for _, p := range projects {
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
		if _, err := os.Stat(filepath.Join(dir, "pom.xml")); err == nil {
			used.maven = true
		}
		for _, script := range []string{"build.gradle.kts", "build.gradle"} {
			if _, err := os.Stat(filepath.Join(dir, script)); err == nil {
				used.gradle = true
			}
		}
	}
	return used
}

func runtime(version string, used buildTools) protocol.Toolchain {
	cache := "maven"
	if !used.maven {
		cache = "gradle"
	}
	return protocol.Toolchain{
		Name: "java", Version: version,
		// GitHub's and Azure's hosted images carry Maven.
		GitHub: &protocol.ActionStep{Uses: "actions/setup-java@v4", Inputs: []protocol.Input{{Name: "distribution", Value: "temurin"}, {Name: "java-version", Value: version}, {Name: "cache", Value: cache}}},
		Azure:  &protocol.ActionStep{Uses: "JavaToolInstaller@0", Inputs: []protocol.Input{{Name: "versionSpec", Value: version}, {Name: "jdkArchitectureOption", Value: "x64"}, {Name: "jdkSourceOption", Value: "PreInstalled"}}},
		// Elsewhere (GitLab, on a Debian image): Temurin and Maven archives.
		Install: "mkdir -p /opt/jdk /opt/maven\n" +
			"curl -fsSL https://api.adoptium.net/v3/binary/latest/" + version + "/ga/linux/x64/jdk/hotspot/normal/eclipse | tar -xz -C /opt/jdk --strip-components=1\n" +
			"curl -fsSL https://archive.apache.org/dist/maven/maven-3/" + MavenVersion + "/binaries/apache-maven-" + MavenVersion + "-bin.tar.gz | tar -xz -C /opt/maven --strip-components=1\n" +
			"export JAVA_HOME=/opt/jdk PATH=/opt/jdk/bin:/opt/maven/bin:$PATH",
		Devcontainer: &protocol.DevcontainerFeature{ID: "ghcr.io/devcontainers/features/java:1", Options: map[string]string{"version": version, "installMaven": "true", "installGradle": strconv.FormatBool(used.gradle)}},
	}
}

// gradle installs Gradle at GradleVersion: setup-gradle on GitHub, the
// distribution elsewhere. Azure runs the same script, so it also prepends
// the path for the later steps (a logging command; plain output elsewhere).
func gradle() protocol.Toolchain {
	bin := "$HOME/gradle-" + GradleVersion + "/bin"
	return protocol.Toolchain{
		Name: "gradle", Version: GradleVersion,
		GitHub: &protocol.ActionStep{Uses: "gradle/actions/setup-gradle@v6", Inputs: []protocol.Input{{Name: "gradle-version", Value: GradleVersion}}},
		Install: "curl -fsSLo /tmp/gradle.zip https://services.gradle.org/distributions/gradle-" + GradleVersion + "-bin.zip\n" +
			"unzip -q -d \"$HOME\" /tmp/gradle.zip && rm /tmp/gradle.zip\n" +
			"export PATH=\"" + bin + ":$PATH\"\n" +
			"echo \"##vso[task.prependpath]" + bin + "\"",
	}
}

// CI returns the JVM contribution: a JDK with Maven, and Gradle when a
// module builds with it, resolving every Maven module's dependencies, and
// the repository credentials when a package is deployed.
func CI(workspaceRoot string, params protocol.CIParams) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	used := usedBuildTools(workspaceRoot, params.Projects)
	result := protocol.CIResult{Toolchains: []protocol.Toolchain{runtime(Version(workspaceRoot), used)}}
	if used.gradle {
		result.Toolchains = append(result.Toolchains, gradle())
	}
	if used.maven {
		result.Setup = []protocol.Step{{Name: "Resolve Maven dependencies", Command: []string{"mvn", "-B", "-q", "dependency:go-offline"}, Blocking: true}}
	}
	for _, p := range params.Projects {
		if p.Kind == protocol.ProjectPackage {
			result.Credentials = append(result.Credentials,
				protocol.Credential{Env: "MAVEN_REPOSITORY_URL", Description: "Maven repository the workspace's packages deploy to", Required: true},
				protocol.Credential{Env: "MAVEN_USERNAME", Description: "Maven repository user", Required: true},
				protocol.Credential{Env: "MAVEN_PASSWORD", Description: "Maven repository password or token", Required: true})
			break
		}
	}
	return result
}

// probe runs a tool for its output; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Output(dir, argv) }

// Doctor checks that a JDK is installed, and Maven or Gradle for the
// modules that build with them.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	if len(params.Projects) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	version := Version(workspaceRoot)
	_, javaErr := probe(workspaceRoot, []string{"java", "-version"})
	findings := []protocol.Finding{
		{Check: "a JDK is installed", OK: javaErr == nil, Detail: "java is not on PATH", Remedy: "install JDK " + version + " (https://adoptium.net)"},
	}
	used := usedBuildTools(workspaceRoot, params.Projects)
	if used.maven {
		_, err := probe(workspaceRoot, []string{"mvn", "-v"})
		findings = append(findings, protocol.Finding{Check: "Maven is installed", OK: err == nil, Detail: "mvn is not on PATH", Remedy: "install Maven " + MavenVersion + " (https://maven.apache.org)"})
	}
	if used.gradle {
		_, err := probe(workspaceRoot, []string{"gradle", "--version"})
		findings = append(findings, protocol.Finding{Check: "Gradle is installed", OK: err == nil, Detail: "gradle is not on PATH", Remedy: "install Gradle " + GradleVersion + " (https://gradle.org/install)"})
	}
	return protocol.DoctorResult{Findings: findings}
}
