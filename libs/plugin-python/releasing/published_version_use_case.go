package releasing

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// distFile matches a distribution's file name on a simple index page and
// captures its version: name-1.2.3.tar.gz, name-1.2.3-py3-none-any.whl.
var distFile = regexp.MustCompile(`>[A-Za-z0-9_.]+?-(\d+(?:\.\d+)*)(?:\.tar\.gz|\.zip|-[^<>]*\.whl)<`)

// Published reads the simple index the project is uploaded to
// (TWINE_REPOSITORY_URL, else PyPI) and returns the newest version among its
// files. A project the index does not have is not published yet.
func Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	registry := os.Getenv("TWINE_REPOSITORY_URL")
	if registry == "" {
		registry = "https://upload.pypi.org/legacy/"
	}
	project := params.Project.Name
	if len(params.Project.Exports) > 0 {
		project = params.Project.Exports[0]
	}
	index := SimpleIndexURL(registry)
	result := protocol.PublishedResult{Registry: index}
	page, err := httpGet(index + normalise(project) + "/")
	if err != nil {
		result.Unreachable = true
		result.Detail = err.Error()
		return result, nil
	}
	for _, m := range distFile.FindAllStringSubmatch(page, -1) {
		if result.Version == "" || newer(m[1], result.Version) {
			result.Version = m[1]
		}
	}
	return result, nil
}

// newer compares dotted numeric versions.
func newer(a, b string) bool {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		var x, y int
		if i < len(as) {
			x, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			y, _ = strconv.Atoi(bs[i])
		}
		if x != y {
			return x > y
		}
	}
	return false
}
