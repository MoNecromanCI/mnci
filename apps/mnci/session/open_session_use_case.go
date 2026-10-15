package session

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/MoNecromanCI/mnci/apps/mnci/gitclient"
	"github.com/MoNecromanCI/mnci/apps/mnci/pluginhost"
	"github.com/MoNecromanCI/mnci/apps/mnci/projectgraph"
	"github.com/MoNecromanCI/mnci/apps/mnci/taskcache"
	"github.com/MoNecromanCI/mnci/apps/mnci/workspace"
)

// Session is an open workspace.
type Session struct {
	Root    string
	Config  workspace.Config
	Git     *gitclient.Git
	Plugins *pluginhost.PluginSet
	Graph   *projectgraph.Graph
	// Files are tracked files plus untracked files that are not ignored.
	Files   []string
	Version string
}

// Options configure Open.
type Options struct {
	Dir     string
	Version string
	// Log receives plugin output; nil means stderr.
	Log io.Writer
}

// Open finds the workspace containing opts.Dir and loads everything a command
// needs. The caller must Close it.
func Open(opts Options) (*Session, error) {
	root, err := workspace.FindRoot(opts.Dir)
	if err != nil {
		return nil, err
	}
	cfg, err := workspace.Load(root)
	if err != nil {
		return nil, err
	}
	git := gitclient.New(root)
	if !git.IsRepository() {
		return nil, errors.New("the workspace is not a git repository; run `git init` at " + root)
	}
	files, err := git.Files()
	if err != nil {
		return nil, err
	}
	plugins, err := pluginhost.Start(pluginhost.Options{WorkspaceRoot: root, Config: cfg, Version: opts.Version, Log: opts.Log})
	if err != nil {
		return nil, err
	}
	graph, err := projectgraph.Discover(files, plugins, cfg)
	if err != nil {
		plugins.Close()
		return nil, err
	}
	return &Session{Root: root, Config: cfg, Git: git, Plugins: plugins, Graph: graph, Files: files, Version: opts.Version}, nil
}

// Storage opens the configured cache.
func (s *Session) Storage() (taskcache.Storage, error) {
	dir := s.Config.Cache.Dir
	if env := os.Getenv("MNCI_CACHE_DIR"); env != "" {
		dir = env
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(s.Root, filepath.FromSlash(dir))
	}
	return taskcache.NewLocalStorage(dir)
}

// Refresh re-reads files and rebuilds the graph, after a command changed the
// workspace (mnci add).
func (s *Session) Refresh() error {
	files, err := s.Git.Files()
	if err != nil {
		return err
	}
	graph, err := projectgraph.Discover(files, s.Plugins, s.Config)
	if err != nil {
		return err
	}
	s.Files, s.Graph = files, graph
	return nil
}

// Close stops the plugins.
func (s *Session) Close() {
	if s != nil && s.Plugins != nil {
		s.Plugins.Close()
	}
}
