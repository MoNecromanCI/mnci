package workspace

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// ErrNotFound is returned when no smith.toml exists at or above a directory.
var ErrNotFound = errors.New("not inside a smith workspace (no smith.toml found); run `smith new <name>` to create one")

// FindRoot walks up from dir to the directory holding smith.toml.
func FindRoot(dir string) (string, error) {
	if env := os.Getenv("SMITH_WORKSPACE_ROOT"); env != "" {
		dir = env
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, FileName)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotFound
		}
		dir = parent
	}
}

// Load reads smith.toml under root, rejects unknown keys and applies defaults.
func Load(root string) (Config, error) {
	raw, err := os.ReadFile(filepath.Join(root, FileName))
	if errors.Is(err, fs.ErrNotExist) {
		return Config{}, ErrNotFound
	}
	if err != nil {
		return Config{}, err
	}
	return Parse(raw)
}

// Parse decodes smith.toml text, rejecting unknown keys so a typo is an error
// rather than a silently ignored setting.
func Parse(raw []byte) (Config, error) {
	var cfg Config
	dec := toml.NewDecoder(bytes.NewReader(raw)).DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		var strict *toml.StrictMissingError
		if errors.As(err, &strict) {
			return cfg, fmt.Errorf("%s: %s", FileName, strings.TrimSpace(strict.String()))
		}
		var decodeErr *toml.DecodeError
		if errors.As(err, &decodeErr) {
			row, col := decodeErr.Position()
			return cfg, fmt.Errorf("%s:%d:%d: %s", FileName, row, col, decodeErr.Error())
		}
		return cfg, fmt.Errorf("%s: %w", FileName, err)
	}
	return cfg, Validate(&cfg)
}

// Validate applies defaults and checks values.
func Validate(cfg *Config) error {
	if cfg.Schema == 0 {
		cfg.Schema = SchemaVersion
	}
	if cfg.Schema > SchemaVersion {
		return fmt.Errorf("%s uses schema %d; this smith reads up to %d, upgrade smith", FileName, cfg.Schema, SchemaVersion)
	}
	if cfg.Schema < SchemaVersion {
		return fmt.Errorf("%s uses schema %d, this smith schema %d: run `smith upgrade`", FileName, cfg.Schema, SchemaVersion)
	}
	if strings.TrimSpace(cfg.Workspace.Name) == "" {
		return fmt.Errorf("%s: [workspace] name is required", FileName)
	}
	for name, st := range cfg.Stacks {
		if strings.TrimSpace(st.Language) == "" {
			return fmt.Errorf("%s: [stacks.%s] needs a language plugin (language = \"node\")", FileName, name)
		}
	}
	if len(cfg.CI.Providers) == 0 {
		cfg.CI.Providers = []string{"github"}
	}
	for _, p := range cfg.CI.Providers {
		if p != "github" && p != "azure" && p != "gitlab" {
			return fmt.Errorf("%s: unknown CI provider %q (use github, azure or gitlab)", FileName, p)
		}
	}
	if cfg.Release.Branch == "" {
		cfg.Release.Branch = "main"
	}
	if cfg.Release.Remote == "" {
		cfg.Release.Remote = "origin"
	}
	if cfg.Cache.Dir == "" {
		cfg.Cache.Dir = ".smith/cache"
	}
	if cfg.Cache.MaxSizeMB == 0 {
		cfg.Cache.MaxSizeMB = 2048
	}
	for name, po := range cfg.Projects {
		if po.Release == nil {
			continue
		}
		if len(po.Release.Platforms) == 0 {
			po.Release.Platforms = append([]string(nil), DefaultPlatforms...)
		}
		if po.Release.AssetName == "" {
			po.Release.AssetName = "{product}_{version}_{os}_{arch}.{ext}"
		}
		for _, p := range po.Release.Platforms {
			if os, arch, ok := strings.Cut(p, "/"); !ok || os == "" || arch == "" {
				return fmt.Errorf("%s: [projects.%s.release] platform %q must be os/arch", FileName, name, p)
			}
		}
	}
	for name, ext := range cfg.Plugins.External {
		if len(ext.Command) > 0 && strings.TrimSpace(ext.Command[0]) == "" {
			return fmt.Errorf("%s: [plugins.external.%s] command starts with an empty program", FileName, name)
		}
	}
	return nil
}
