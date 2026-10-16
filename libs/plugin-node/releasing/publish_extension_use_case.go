package releasing

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// vsceFlags let vsce package a staged build: its dependencies are bundled,
// and a missing LICENSE or repository is the author's choice, not a prompt.
var vsceFlags = []string{"--no-dependencies", "--skip-license", "--allow-missing-repository"}

// ExtensionManifest is what marks a package as a VS Code extension.
type ExtensionManifest struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Publisher string `json:"publisher"`
	Engines   struct {
		VSCode string `json:"vscode"`
	} `json:"engines"`
}

// Extension reads dir's package.json when it is a VS Code extension (it
// names an engines.vscode); ok is false for any other package.
func Extension(dir string) (ExtensionManifest, bool) {
	raw, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return ExtensionManifest{}, false
	}
	var m ExtensionManifest
	if json.Unmarshal(raw, &m) != nil || m.Engines.VSCode == "" {
		return ExtensionManifest{}, false
	}
	return m, true
}

// PackExtension packages a staged extension into drop as
// <name>-<version>.vsix and returns the file.
func PackExtension(staged, drop string) (string, error) {
	m, _ := Extension(staged)
	out := filepath.Join(drop, m.Name+"-"+m.Version+".vsix")
	argv := append([]string{"npx", "--no-install", "vsce", "package"}, vsceFlags...)
	return out, npm.run(staged, append(argv, "--out", out))
}

// publishExtension publishes a VS Code extension to the Marketplace with
// the token in VSCE_PAT, from its staged build. vsce skips a version the
// Marketplace already has.
func publishExtension(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	result := protocol.PublishResult{Registry: "https://marketplace.visualstudio.com"}
	if params.DryRun {
		result.Skipped = "dry run"
		return result, nil
	}
	if os.Getenv("VSCE_PAT") == "" {
		return result, errors.New("set VSCE_PAT to a Marketplace token to publish " + params.Project.Name)
	}
	staged, err := Stage(workspaceRoot, params.Project.Root)
	if err != nil {
		return result, err
	}
	argv := append([]string{"npx", "--no-install", "vsce", "publish"}, vsceFlags...)
	if err := npm.run(staged, append(argv, "--skip-duplicate")); err != nil {
		return result, err
	}
	result.Published = true
	return result, nil
}

// publishedExtension asks the Marketplace for an extension's newest
// version; one it does not know is not published yet.
func publishedExtension(workspaceRoot string, m ExtensionManifest) protocol.PublishedResult {
	result := protocol.PublishedResult{Registry: "https://marketplace.visualstudio.com"}
	out, errOut, err := capture(workspaceRoot, []string{"npx", "--no-install", "vsce", "show", m.Publisher + "." + m.Name, "--json"})
	if err != nil {
		if !strings.Contains(strings.ToLower(errOut), "not found") {
			result.Unreachable, result.Detail = true, firstLine(errOut, err.Error())
		}
		return result
	}
	var shown struct {
		Versions []struct {
			Version string `json:"version"`
		} `json:"versions"`
	}
	if json.Unmarshal([]byte(out), &shown) != nil {
		result.Unreachable, result.Detail = true, "vsce show returned no JSON"
		return result
	}
	if len(shown.Versions) > 0 {
		result.Version = shown.Versions[0].Version
	}
	return result
}
