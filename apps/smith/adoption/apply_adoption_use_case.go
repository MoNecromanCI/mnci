package adoption

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/MoNecromanCI/mnci/apps/mnci/cipipeline"
	"github.com/MoNecromanCI/mnci/apps/mnci/gitclient"
	"github.com/MoNecromanCI/mnci/apps/mnci/scaffold"
	"github.com/MoNecromanCI/mnci/apps/mnci/session"
)

// ErrDirty is a step asked for with uncommitted changes.
var ErrDirty = errors.New("commit or stash your changes first: the adoption steps change the repository, and a clean tree keeps them reviewable")

// CreateBaselines tags, locally, the commits the lineage checks found under
// an old name with the project's current tag name; nothing is pushed.
func CreateBaselines(r Report, out io.Writer) (int, error) {
	git := gitclient.New(r.Root)
	existing, err := git.Tags("")
	if err != nil {
		return 0, err
	}
	made := 0
	for _, b := range r.Blockers {
		if b.Baseline == nil || slices.Contains(existing, b.Baseline.Name) {
			continue
		}
		if _, err := git.Run("tag", b.Baseline.Name, b.Baseline.Target); err != nil {
			return made, fmt.Errorf("tag %s: %w", b.Baseline.Name, err)
		}
		fmt.Fprintf(out, "  tagged %s (%s); push it with git push origin %s\n", b.Baseline.Name, strings.TrimSuffix(b.Baseline.Target, "^{commit}"), b.Baseline.Name)
		made++
	}
	return made, nil
}

// Init makes the repository a workspace: a pipeline file of the user's at a
// path mnci owns is kept beside it as .pre-mnci, mnci.toml and the root files
// mnci needs are written (theirs kept), the pipelines are generated for the
// projects found, and it all becomes one commit.
func Init(r Report, version string, out io.Writer) error {
	if !r.Adoptable() {
		return errors.New("something blocks adoption; see mnci adopt")
	}
	if !r.Clean {
		return ErrDirty
	}
	git := gitclient.New(r.Root)
	for _, path := range []string{cipipeline.GitHubPath, cipipeline.AzurePath, cipipeline.DevcontainerPath} {
		full := filepath.Join(r.Root, filepath.FromSlash(path))
		body, err := os.ReadFile(full)
		if err != nil || strings.Contains(string(body), GeneratedMarker) {
			continue
		}
		if _, err := git.Run("mv", path, path+".pre-mnci"); err != nil {
			return fmt.Errorf("keep %s: %w", path, err)
		}
		fmt.Fprintf(out, "  kept your %s as %s.pre-mnci\n", path, path)
	}
	if _, err := scaffold.NewWorkspace(scaffold.NewOptions{Name: r.Name, Dir: r.Root, Scope: r.Scope, Providers: r.Providers, Version: version, Out: out, NoCommit: true}); err != nil {
		return err
	}
	s, err := session.Open(session.Options{Dir: r.Root, Version: version, Log: out})
	if err != nil {
		return err
	}
	defer s.Close()
	contributions, err := cipipeline.Collect(s)
	if err != nil {
		return err
	}
	if _, _, err := cipipeline.Write(s.Root, s.Config, cipipeline.NewModel(contributions, version, s.Config.Release.Branch)); err != nil {
		return err
	}
	if err := git.CommitAll("chore: adopt mnci"); err != nil {
		fmt.Fprintf(out, "  not committed (%v); commit when ready\n", strings.SplitN(err.Error(), "\n", 2)[0])
		return nil
	}
	fmt.Fprintln(out, "  committed: chore: adopt mnci")
	return nil
}
