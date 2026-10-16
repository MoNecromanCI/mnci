package discovery

import (
	"encoding/json"
	"sort"
	"strings"
)

// Manifest is what mnci reads from a composer.json.
type Manifest struct {
	Name, Version, Type string
	// Kind and Stack are extra.mnci's.
	Kind, Stack string
	// Bin are the package's executables.
	Bin []string
	// Requires are the names of every requirement, require-dev too, without
	// the platform ones (php, ext-*).
	Requires []string
}

type manifestJSON struct {
	Name       string            `json:"name"`
	Version    string            `json:"version"`
	Type       string            `json:"type"`
	Bin        []string          `json:"bin"`
	Require    map[string]string `json:"require"`
	RequireDev map[string]string `json:"require-dev"`
	Extra      struct {
		Mnci struct {
			Kind  string `json:"kind"`
			Stack string `json:"stack"`
		} `json:"mnci"`
	} `json:"extra"`
}

// platform reports a requirement on PHP itself or an extension.
func platform(name string) bool {
	return name == "php" || !strings.Contains(name, "/")
}

// ParseManifest reads a composer.json.
func ParseManifest(src []byte) (Manifest, error) {
	var j manifestJSON
	if err := json.Unmarshal(src, &j); err != nil {
		return Manifest{}, err
	}
	m := Manifest{Name: j.Name, Version: j.Version, Type: j.Type, Bin: j.Bin, Kind: j.Extra.Mnci.Kind, Stack: j.Extra.Mnci.Stack}
	for _, reqs := range []map[string]string{j.Require, j.RequireDev} {
		for name := range reqs {
			if !platform(name) {
				m.Requires = append(m.Requires, name)
			}
		}
	}
	sort.Strings(m.Requires)
	return m, nil
}
