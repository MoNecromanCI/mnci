package cipipeline

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/MoNecromanCI/mnci/apps/mnci/session"
	"github.com/MoNecromanCI/mnci/apps/mnci/workspace"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Collect asks every plugin what CI needs for its projects.
func Collect(s *session.Session) (Contributions, error) {
	var c Contributions
	for _, client := range s.Plugins.Clients {
		if !client.Info.Capabilities.CI {
			continue
		}
		var projects []protocol.Project
		for _, p := range s.Graph.Projects {
			if p.ServedBy(client.Info.Plugin.Name) {
				projects = append(projects, p.Project)
			}
		}
		if len(projects) == 0 {
			continue
		}
		var r protocol.CIResult
		if err := client.Call(protocol.MethodCI, protocol.CIParams{Projects: projects}, &r, time.Minute); err != nil {
			return c, fmt.Errorf("plugin %s: %w", client.Info.Plugin.Name, err)
		}
		c.Add(r)
	}
	for _, p := range s.Graph.Projects {
		if _, ok := p.Tasks[PackTask]; ok {
			c.Pack = true
		}
	}
	if s.Config.Release.GitHubReleases {
		c.Add(protocol.CIResult{Credentials: []protocol.Credential{{Env: "GITHUB_TOKEN", Description: "creates GitHub Releases and uploads their assets", Required: true}}})
	}
	return c, nil
}

// Rendered maps each configured provider's file to its content.
func Rendered(cfg workspace.Config, m Model) map[string]string {
	out := map[string]string{}
	for _, p := range cfg.CI.Providers {
		switch p {
		case "github":
			out[GitHubPath] = RenderGitHub(m)
		case "azure":
			out[AzurePath] = RenderAzure(m)
		}
	}
	return out
}

// Write renders and writes every pipeline, keeping slot contents. It returns
// the files that changed and any slot that could not be kept.
func Write(root string, cfg workspace.Config, m Model) (changed, lost []string, err error) {
	for path, content := range Rendered(cfg, m) {
		full := filepath.Join(root, filepath.FromSlash(path))
		old, readErr := os.ReadFile(full)
		if readErr != nil && !errors.Is(readErr, fs.ErrNotExist) {
			return changed, lost, readErr
		}
		content, missing := KeepSlots(string(old), content)
		for _, slot := range missing {
			lost = append(lost, path+": "+slot)
		}
		if string(old) == content {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return changed, lost, err
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			return changed, lost, err
		}
		changed = append(changed, path)
	}
	return changed, lost, nil
}

// Drift lists pipelines whose committed content differs from what the
// workspace would generate now.
func Drift(root string, cfg workspace.Config, m Model) []string {
	var drift []string
	for path, content := range Rendered(cfg, m) {
		old, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			drift = append(drift, path+" is missing")
			continue
		}
		if kept, _ := KeepSlots(string(old), content); kept != string(old) {
			drift = append(drift, path+" is out of date")
		}
	}
	return drift
}

// Regenerate collects contributions and rewrites the pipelines of an open
// workspace.
func Regenerate(s *session.Session) (changed, lost []string, err error) {
	c, err := Collect(s)
	if err != nil {
		return nil, nil, err
	}
	return Write(s.Root, s.Config, NewModel(c, s.Version, s.Config.Release.Branch))
}
