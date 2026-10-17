package splitrepo

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
)

// Request is one release to mirror.
type Request struct {
	// Prefix is the project's directory, workspace-relative.
	Prefix string
	// Ref is the monorepo's release tag the split is taken from.
	Ref string
	// Repo is the mirror, owner/repo on GitHub.
	Repo string
	// URL overrides the mirror's address (tests push to a local repository).
	URL     string
	Version string
	Token   string
	DryRun  bool
}

// Result says what happened.
type Result struct {
	// Mirror is owner/repo@tag.
	Mirror string
	// Skipped says why nothing was pushed (the mirror already has the tag).
	Skipped string
}

// lookPath finds splitsh-lite; tests replace it.
var lookPath = exec.LookPath

// Mirror splits the project's directory out of the history at the release
// tag and pushes it to the mirror's main branch and tag. A mirror that
// already has the tag is left alone, so a re-run release resumes. The token
// travels in git's environment-set config, never in an argument, so no error
// or log line can show it.
func Mirror(g *gitclient.Git, req Request) (Result, error) {
	tag := Tag(req.Version)
	res := Result{Mirror: req.Repo + "@" + tag}
	if req.DryRun {
		return res, nil
	}
	if req.Token == "" && req.URL == "" {
		return res, fmt.Errorf("%s is not set: %s", TokenEnv, Credential(req.Repo).Description)
	}
	url := req.URL
	if url == "" {
		url = "https://github.com/" + req.Repo + ".git"
	}
	env := authEnv(req.Token)
	remote, err := g.RunWithEnv(env, "ls-remote", "--tags", url, "refs/tags/"+tag)
	if err != nil {
		return res, fmt.Errorf("reading %s: %w", req.Repo, err)
	}
	if strings.TrimSpace(remote) != "" {
		res.Skipped = req.Repo + " already has " + tag
		return res, nil
	}
	commit, err := split(g, req.Prefix, req.Ref)
	if err != nil {
		return res, err
	}
	if _, err := g.RunWithEnv(env, "push", "--no-verify", url, commit+":refs/heads/"+Branch, commit+":refs/tags/"+tag); err != nil {
		return res, fmt.Errorf("pushing to %s: %w (the monorepo is tagged; re-run the release to push the mirror)", req.Repo, err)
	}
	return res, nil
}

// split returns the commit holding the history of prefix alone, at ref.
// splitsh-lite, when installed, does it much faster on a long history.
func split(g *gitclient.Git, prefix, ref string) (string, error) {
	if path, err := lookPath("splitsh-lite"); err == nil {
		cmd := exec.Command(path, "--prefix="+prefix, "--origin="+ref)
		cmd.Dir = g.Dir
		out, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("splitsh-lite --prefix=%s: %w", prefix, err)
		}
		return strings.TrimSpace(string(out)), nil
	}
	out, err := g.Run("subtree", "split", "--prefix="+prefix, ref)
	if err != nil {
		return "", fmt.Errorf("splitting %s: %w", prefix, err)
	}
	lines := strings.Fields(out)
	if len(lines) == 0 {
		return "", errors.New("git subtree split printed no commit for " + prefix)
	}
	return lines[len(lines)-1], nil
}

// authEnv authenticates to github.com with the token as actions/checkout
// does, first emptying the header list, so a checkout's own GITHUB_TOKEN
// header is not sent beside it.
func authEnv(token string) []string {
	if token == "" {
		return nil
	}
	header := "AUTHORIZATION: basic " + base64.StdEncoding.EncodeToString([]byte("x-access-token:"+token))
	key := "http.https://github.com/.extraheader"
	return []string{
		"GIT_CONFIG_COUNT=2",
		"GIT_CONFIG_KEY_0=" + key, "GIT_CONFIG_VALUE_0=",
		"GIT_CONFIG_KEY_1=" + key, "GIT_CONFIG_VALUE_1=" + header,
		"GIT_TERMINAL_PROMPT=0",
	}
}

// Token is the mirror token from the environment.
func Token() string { return os.Getenv(TokenEnv) }
