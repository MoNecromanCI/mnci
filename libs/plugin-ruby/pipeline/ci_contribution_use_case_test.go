package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestCI(t *testing.T) {
	ws := t.TempDir()
	res := CI(ws, protocol.CIParams{Projects: []protocol.Project{{Kind: protocol.ProjectPackage}}})
	if res.Toolchains[0].Version != RubyVersion || res.Setup[0].Command[1] != "install" || res.Credentials[0].Env != "GEM_HOST_API_KEY" {
		t.Fatalf("%+v", res)
	}
	_ = os.WriteFile(filepath.Join(ws, ".ruby-version"), []byte("ruby-3.3.6\n"), 0o644)
	if v := Version(ws); v != "3.3.6" {
		t.Fatalf(".ruby-version pins %s", v)
	}
}

func TestDoctor(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	probe = func(_ string, argv []string) (string, error) {
		if argv[0] == "bundle" {
			return "", errors.New("missing")
		}
		return "", nil
	}
	res := Doctor(t.TempDir(), protocol.DoctorParams{Projects: []protocol.Project{{}}})
	if len(res.Findings) != 2 || !res.Findings[0].OK || res.Findings[1].OK {
		t.Fatalf("%+v", res)
	}
}
