package release

import (
	"fmt"
	"regexp"
	"strconv"
)

var semverPattern = regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-([0-9A-Za-z.-]+))?$`)

// Version is a semantic version.
type Version struct {
	Major, Minor, Patch int
	Pre                 string
}

// ParseVersion parses "1.2.3", "v1.2.3" or "1.2.3-rc.1".
func ParseVersion(s string) (Version, error) {
	m := semverPattern.FindStringSubmatch(s)
	if m == nil {
		return Version{}, fmt.Errorf("%q is not a semantic version", s)
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	return Version{major, minor, patch, m[4]}, nil
}

func (v Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Pre != "" {
		s += "-" + v.Pre
	}
	return s
}

// Less orders versions; a prerelease sorts before its release.
func (v Version) Less(o Version) bool {
	if v.Major != o.Major {
		return v.Major < o.Major
	}
	if v.Minor != o.Minor {
		return v.Minor < o.Minor
	}
	if v.Patch != o.Patch {
		return v.Patch < o.Patch
	}
	return v.Pre != "" && (o.Pre == "" || v.Pre < o.Pre)
}

// Bump applies a level. Before 1.0 a breaking change is a minor bump, as the
// semver spec allows for initial development.
func (v Version) Bump(l Level) Version {
	if v.Major == 0 && l == Major {
		l = Minor
	}
	switch l {
	case Major:
		return Version{Major: v.Major + 1}
	case Minor:
		return Version{Major: v.Major, Minor: v.Minor + 1}
	case Patch:
		if v.Pre != "" {
			return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch}
		}
		return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
	}
	return v
}
