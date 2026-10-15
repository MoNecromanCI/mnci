package release

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/MoNecromanCI/mnci/apps/mnci/githubrelease"
	"github.com/MoNecromanCI/mnci/apps/mnci/session"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Preflight refuses a real release that would fail halfway or corrupt history:
// a shallow clone, the wrong branch, uncommitted changes (manifests are
// stamped and restored), or a missing publish credential.
func Preflight(s *session.Session, plan Plan) error {
	var problems []string
	if s.Git.IsShallow() {
		problems = append(problems, "the clone is shallow, so versions cannot be computed from history: fetch with full depth (actions/checkout fetch-depth: 0)")
	}
	if branch := s.Git.CurrentBranch(); branch != "" && branch != s.Config.Release.Branch && os.Getenv("MNCI_RELEASE_ANY_BRANCH") == "" {
		problems = append(problems, fmt.Sprintf("releases run from %s, not %s", s.Config.Release.Branch, branch))
	}
	if out, err := s.Git.Run("status", "--porcelain", "--untracked-files=no"); err == nil && out != "" {
		problems = append(problems, "the working tree has uncommitted changes; commit or stash them first")
	}
	if s.Config.Release.GitHubReleases && len(plan.Items) > 0 {
		if _, err := githubrelease.FromEnvironment(); err != nil {
			problems = append(problems, err.Error())
		}
		if _, err := Repository(s); err != nil {
			problems = append(problems, err.Error())
		}
	}
	plugins := map[string]bool{}
	for _, item := range plan.Items {
		plugins[item.Plugin] = true
	}
	for name := range plugins {
		var projects []protocol.Project
		for _, p := range s.Graph.Projects {
			if p.Plugin == name {
				projects = append(projects, p.Project)
			}
		}
		var ci protocol.CIResult
		if ok, err := s.Plugins.Call(name, protocol.MethodCI, protocol.CIParams{Projects: projects}, &ci, time.Minute); err != nil || !ok {
			continue
		}
		for _, c := range ci.Credentials {
			if c.Required && os.Getenv(c.Env) == "" && !alternative(c.Env) {
				problems = append(problems, fmt.Sprintf("%s is not set: %s", c.Env, c.Description))
			}
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("cannot release:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return nil
}

// alternative accepts the credential forms each tool also understands.
func alternative(env string) bool {
	switch env {
	case "PYPI_TOKEN":
		return os.Getenv("TWINE_PASSWORD") != ""
	case "NODE_AUTH_TOKEN":
		return os.Getenv("NPM_TOKEN") != "" || os.Getenv("MNCI_NPM_ALREADY_AUTHENTICATED") != ""
	}
	return false
}
