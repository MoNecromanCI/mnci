package projectgraph

import (
	"fmt"
	"slices"

	"github.com/MoNecromanCI/mnci/apps/mnci/workspace"
)

// ConventionTargets picks the projects a convention applies to: those of the
// languages it supports (or the subset mnci.toml names) that are not
// excluded. A language the convention does not support is an error, so a
// typo cannot silently switch it off.
func ConventionTargets(name string, projects []*Project, supported []string, cfg workspace.ConventionConfig) ([]*Project, error) {
	languages := supported
	if len(cfg.Languages) > 0 {
		for _, l := range cfg.Languages {
			if !slices.Contains(supported, l) {
				return nil, fmt.Errorf("mnci.toml [conventions.%s]: it does not apply to %q; it supports %v", name, l, supported)
			}
		}
		languages = cfg.Languages
	}
	var out []*Project
	for _, p := range projects {
		if slices.Contains(languages, p.Language) && !slices.Contains(cfg.Exclude, p.Name) {
			out = append(out, p)
		}
	}
	return out, nil
}
