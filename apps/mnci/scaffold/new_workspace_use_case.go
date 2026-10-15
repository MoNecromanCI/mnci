package scaffold

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/MoNecromanCI/mnci/apps/mnci/cipipeline"
	"github.com/MoNecromanCI/mnci/apps/mnci/gitclient"
	"github.com/MoNecromanCI/mnci/apps/mnci/workspace"
)

// NewOptions configure a new workspace.
type NewOptions struct {
	Name string
	// Dir is where to create it; default ./<name>. It may already exist and
	// hold files mnci does not write (a fresh clone with a LICENSE).
	Dir          string
	Scope        string
	ModulePrefix string
	Providers    []string
	Version      string
	Out          io.Writer
}

// NewWorkspace creates a workspace: mnci.toml, the root files, a git
// repository, the pipelines, and a first commit when git has an identity.
func NewWorkspace(opts NewOptions) (string, error) {
	if err := ValidateName("workspace", opts.Name); err != nil {
		return "", err
	}
	dir := opts.Dir
	if dir == "" {
		dir = opts.Name
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(dir, workspace.FileName)); err == nil {
		return "", fmt.Errorf("%s already holds an mnci workspace", dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	git := gitclient.New(dir)
	if !git.IsRepository() {
		if err := git.Init(); err != nil {
			return "", err
		}
	}
	if opts.Scope == "" {
		opts.Scope = "@" + opts.Name
	}
	if !strings.HasPrefix(opts.Scope, "@") {
		opts.Scope = "@" + opts.Scope
	}
	if opts.ModulePrefix == "" {
		opts.ModulePrefix = modulePrefixFromRemote(git, opts.Name)
	}
	if len(opts.Providers) == 0 {
		opts.Providers = []string{"github"}
	}
	quoted := make([]string, len(opts.Providers))
	for i, p := range opts.Providers {
		quoted[i] = fmt.Sprintf("%q", p)
	}
	config := fmt.Sprintf(configTemplate, opts.Name, opts.Scope, opts.ModulePrefix, strings.Join(quoted, ", "))
	cfg, err := workspace.Parse([]byte(config))
	if err != nil {
		return "", err
	}
	files := map[string]string{
		workspace.FileName: config,
		".gitignore":       gitignore,
		".gitattributes":   gitattributes,
		"README.md":        fmt.Sprintf(readme, opts.Name),
	}
	for name, body := range files {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil && name != workspace.FileName {
			fmt.Fprintf(opts.Out, "  kept your %s\n", name)
			continue
		} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			return "", err
		}
		fmt.Fprintf(opts.Out, "  wrote %s\n", name)
	}
	written, _, err := cipipeline.Write(dir, cfg, cipipeline.NewModel(cipipeline.Contributions{}, opts.Version, cfg.Release.Branch))
	if err != nil {
		return "", err
	}
	for _, f := range written {
		fmt.Fprintf(opts.Out, "  wrote %s\n", f)
	}
	if err := git.CommitAll("chore: create the " + opts.Name + " workspace with mnci"); err != nil {
		fmt.Fprintf(opts.Out, "  not committed (%v); commit when ready\n", firstLine(err.Error()))
	} else {
		fmt.Fprintln(opts.Out, "  committed")
	}
	return dir, nil
}

// modulePrefixFromRemote derives github.com/<owner>/<repo> from the origin
// remote, falling back to example.com/<name>.
func modulePrefixFromRemote(git *gitclient.Git, name string) string {
	url, err := git.RemoteURL("origin")
	if err != nil || url == "" {
		return "example.com/" + name
	}
	url = strings.TrimSuffix(strings.TrimSpace(url), ".git")
	for _, prefix := range []string{"https://", "http://", "ssh://", "git@"} {
		url = strings.TrimPrefix(url, prefix)
	}
	url = strings.Replace(url, ":", "/", 1)
	if i := strings.Index(url, "@"); i >= 0 {
		url = url[i+1:]
	}
	return strings.ToLower(url[:1]) + url[1:]
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}
