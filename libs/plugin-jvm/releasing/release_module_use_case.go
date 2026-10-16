package releasing

import (
	"encoding/xml"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/toolrun"
)

// MavenRunner runs mvn; tests replace it.
var MavenRunner = func(dir string, argv []string) error { return toolrun.Run(dir, argv) }

// skipped are directories no workspace pom lives in.
var skipped = map[string]bool{".git": true, protocol.BuildDir: true, "node_modules": true, ".mnci": true}

// WriteVersion records a release's version in the module's pom and in every
// workspace pom that depends on the module with a version.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	result := protocol.VersionResult{Files: []string{}}
	own := path.Join(params.Project.Root, "pom.xml")
	group, artifact, _ := strings.Cut(firstExport(params.Project), ":")
	err := filepath.WalkDir(workspaceRoot, func(full string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipped[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "pom.xml" {
			return nil
		}
		rel, _ := filepath.Rel(workspaceRoot, full)
		rel = filepath.ToSlash(rel)
		src, err := os.ReadFile(full)
		if err != nil {
			return err
		}
		out, changed := string(src), false
		if rel == own {
			out, changed = SetVersion(out, params.Version)
		}
		if artifact != "" {
			var dep bool
			out, dep = SetDependencyVersion(out, group, artifact, params.Version)
			changed = changed || dep
		}
		if !changed {
			return nil
		}
		result.Files = append(result.Files, rel)
		if params.DryRun {
			return nil
		}
		return os.WriteFile(full, []byte(out), 0o644)
	})
	return result, err
}

func firstExport(p protocol.Project) string {
	if len(p.Exports) > 0 {
		return p.Exports[0]
	}
	return ""
}

// MavenCentral is where versions are read when MAVEN_REPOSITORY_URL is not
// set; tests replace it.
var MavenCentral = "https://repo1.maven.org/maven2"

// settings is the Maven settings Publish writes: the repository's
// credentials, read from the environment by Maven itself.
const settings = `<settings>
  <servers>
    <server>
      <id>mnci</id>
      <username>${env.MAVEN_USERNAME}</username>
      <password>${env.MAVEN_PASSWORD}</password>
    </server>
  </servers>
</settings>
`

// Publish deploys a package to MAVEN_REPOSITORY_URL with the credentials in
// MAVEN_USERNAME and MAVEN_PASSWORD. Apps and libraries release by their tag.
func Publish(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	if params.Assets != nil {
		return protocol.PublishResult{Skipped: "JVM release assets are not built yet: the app is released by its tag"}, nil
	}
	if params.Project.Kind != protocol.ProjectPackage {
		return protocol.PublishResult{Skipped: "released by its tag"}, nil
	}
	repository := os.Getenv("MAVEN_REPOSITORY_URL")
	result := protocol.PublishResult{Registry: repository}
	if params.DryRun {
		result.Skipped = "dry run"
		return result, nil
	}
	if repository == "" {
		return result, fmt.Errorf("set MAVEN_REPOSITORY_URL to the repository %s deploys to", params.Project.Name)
	}
	dir := filepath.Join(workspaceRoot, protocol.BuildDir, "maven")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return result, err
	}
	settingsFile := filepath.Join(dir, "settings.xml")
	if err := os.WriteFile(settingsFile, []byte(settings), 0o600); err != nil {
		return result, err
	}
	argv := []string{"mvn", "-B", "-q", "-pl", params.Project.Root, "deploy", "-DskipTests", "-s", settingsFile, "-DaltDeploymentRepository=mnci::" + repository}
	if err := MavenRunner(workspaceRoot, argv); err != nil {
		return result, err
	}
	result.Published = true
	return result, nil
}

// Published reads the newest release from the repository's
// maven-metadata.xml (MAVEN_REPOSITORY_URL, else Maven Central).
func Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	group, artifact, ok := strings.Cut(firstExport(params.Project), ":")
	if params.Project.Kind != protocol.ProjectPackage || !ok {
		return protocol.PublishedResult{}, nil
	}
	repository := strings.TrimSuffix(os.Getenv("MAVEN_REPOSITORY_URL"), "/")
	if repository == "" {
		repository = MavenCentral
	}
	result := protocol.PublishedResult{Registry: repository}
	url := repository + "/" + strings.ReplaceAll(group, ".", "/") + "/" + artifact + "/maven-metadata.xml"
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Get(url)
	if err != nil {
		result.Unreachable, result.Detail = true, err.Error()
		return result, nil
	}
	defer func() { _ = resp.Body.Close() }()
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return result, nil
	case resp.StatusCode != http.StatusOK:
		result.Unreachable, result.Detail = true, resp.Status
		return result, nil
	}
	var metadata struct {
		Release string `xml:"versioning>release"`
	}
	if err := xml.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		result.Unreachable, result.Detail = true, err.Error()
		return result, nil
	}
	result.Version = metadata.Release
	return result, nil
}
