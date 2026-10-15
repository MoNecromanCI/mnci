package releasing

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/toolrun"
)

// ChecksumsFile is attached beside the archives; the installers verify
// downloads against it.
const ChecksumsFile = "checksums.txt"

// Publish releases a Go project. A module is published by its tag, so there
// is nothing to upload; an app with assets is cross-compiled for every
// platform with the version stamped in, and each binary is zipped.
func Publish(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	if params.Assets == nil {
		return protocol.PublishResult{Skipped: "published by its tag"}, nil
	}
	return BuildAssets(workspaceRoot, params)
}

// BuildAssets builds, zips and checksums one binary per platform.
func BuildAssets(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	spec := *params.Assets
	p := params.Project
	if len(spec.Platforms) == 0 {
		return protocol.PublishResult{}, fmt.Errorf("%s: no platforms to build assets for", p.Name)
	}
	variable := spec.VersionVariable
	if variable == "" {
		variable = "main.version"
	}
	outDir := filepath.Join(workspaceRoot, filepath.FromSlash(spec.OutDir))
	var assets []string
	var sums strings.Builder
	if !params.DryRun {
		if err := os.RemoveAll(outDir); err != nil {
			return protocol.PublishResult{}, err
		}
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return protocol.PublishResult{}, err
		}
	}
	for _, platform := range spec.Platforms {
		goos, goarch, ok := strings.Cut(platform, "/")
		if !ok || goos == "" || goarch == "" {
			return protocol.PublishResult{}, fmt.Errorf("%s: platform %q is not os/arch", p.Name, platform)
		}
		name, err := AssetName(spec, p.Name, params.Version, goos, goarch, "zip")
		if err != nil {
			return protocol.PublishResult{}, err
		}
		assets = append(assets, path.Join(spec.OutDir, name))
		if params.DryRun {
			continue
		}
		binary := p.Name
		if goos == "windows" {
			binary += ".exe"
		}
		built := filepath.Join(outDir, goos+"-"+goarch, binary)
		argv := []string{"go", "build", "-trimpath", "-ldflags", "-s -w -X " + variable + "=" + params.Version, "-o", built, "."}
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
		if err := toolrun.Run(dir, argv, "CGO_ENABLED=0", "GOOS="+goos, "GOARCH="+goarch); err != nil {
			return protocol.PublishResult{}, fmt.Errorf("building %s: %w", platform, err)
		}
		archive := filepath.Join(outDir, name)
		if err := zipFile(archive, built, binary); err != nil {
			return protocol.PublishResult{}, err
		}
		sum, err := digest(archive)
		if err != nil {
			return protocol.PublishResult{}, err
		}
		fmt.Fprintf(&sums, "%s  %s\n", sum, name)
		if err := os.RemoveAll(filepath.Dir(built)); err != nil {
			return protocol.PublishResult{}, err
		}
	}
	assets = append(assets, path.Join(spec.OutDir, ChecksumsFile))
	if !params.DryRun {
		if err := os.WriteFile(filepath.Join(outDir, ChecksumsFile), []byte(sums.String()), 0o644); err != nil {
			return protocol.PublishResult{}, err
		}
	}
	return protocol.PublishResult{Published: !params.DryRun, Assets: assets}, nil
}

// zipFile writes an archive holding one executable at its root.
func zipFile(archive, file, name string) error {
	out, err := os.Create(archive)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(out)
	in, err := os.Open(file)
	if err != nil {
		_ = out.Close()
		return err
	}
	defer func() { _ = in.Close() }()
	header := &zip.FileHeader{Name: name, Method: zip.Deflate}
	header.SetMode(0o755)
	w, err := zw.CreateHeader(header)
	if err == nil {
		_, err = io.Copy(w, in)
	}
	if cerr := zw.Close(); err == nil {
		err = cerr
	}
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	return err
}

func digest(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
