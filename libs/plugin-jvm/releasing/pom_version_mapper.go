package releasing

import (
	"regexp"
	"strings"
)

var (
	parentBlock     = regexp.MustCompile(`(?s)<parent>.*?</parent>`)
	versionElement  = regexp.MustCompile(`<version>[^<]*</version>`)
	dependencyBlock = regexp.MustCompile(`(?s)<dependency>.*?</dependency>`)
	// The sections after a pom's own coordinates.
	sections = []string{"<properties>", "<dependencies>", "<dependencyManagement>", "<build>", "<modules>", "<profiles>"}
)

// SetVersion writes version as the pom's own <version> (not its parent's),
// adding one after its <artifactId> when it inherits the parent's.
func SetVersion(src, version string) (string, bool) {
	start := 0
	if loc := parentBlock.FindStringIndex(src); loc != nil {
		start = loc[1]
	}
	end := len(src)
	for _, s := range sections {
		if i := strings.Index(src[start:], s); i >= 0 && start+i < end {
			end = start + i
		}
	}
	element := "<version>" + version + "</version>"
	if loc := versionElement.FindStringIndex(src[start:end]); loc != nil {
		out := src[:start+loc[0]] + element + src[start+loc[1]:]
		return out, out != src
	}
	i := strings.Index(src[start:end], "</artifactId>")
	if i < 0 {
		return src, false
	}
	at := start + i + len("</artifactId>")
	line := src[strings.LastIndex(src[:at], "\n")+1 : at]
	indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
	return src[:at] + "\n" + indent + element + src[at:], true
}

// SetDependencyVersion writes version into every <dependency> on
// group:artifact that names a version.
func SetDependencyVersion(src, group, artifact, version string) (string, bool) {
	out := dependencyBlock.ReplaceAllStringFunc(src, func(dep string) string {
		if !strings.Contains(dep, "<artifactId>"+artifact+"</artifactId>") {
			return dep
		}
		if !strings.Contains(dep, "<groupId>"+group+"</groupId>") && !strings.Contains(dep, "<groupId>${project.groupId}</groupId>") {
			return dep
		}
		return versionElement.ReplaceAllString(dep, "<version>"+version+"</version>")
	})
	return out, out != src
}
