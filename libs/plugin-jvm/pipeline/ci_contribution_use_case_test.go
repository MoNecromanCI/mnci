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
	if res.Toolchains[0].Version != JavaVersion || res.Setup[0].Command[3] != "dependency:go-offline" || len(res.Credentials) != 3 {
		t.Fatalf("%+v", res)
	}
	_ = os.WriteFile(filepath.Join(ws, "pom.xml"), []byte("<properties><maven.compiler.release>17</maven.compiler.release></properties>"), 0o644)
	if v := Version(ws); v != "17" {
		t.Fatalf("the root pom pins %s", v)
	}
	_ = os.WriteFile(filepath.Join(ws, ".java-version"), []byte("25.0.1\n"), 0o644)
	if v := Version(ws); v != "25" {
		t.Fatalf(".java-version pins %s", v)
	}
}

func TestDoctor(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	probe = func(_ string, argv []string) (string, error) {
		if argv[0] == "mvn" {
			return "", errors.New("missing")
		}
		return "", nil
	}
	res := Doctor(t.TempDir(), protocol.DoctorParams{Projects: []protocol.Project{{}}})
	if len(res.Findings) != 2 || !res.Findings[0].OK || res.Findings[1].OK {
		t.Fatalf("%+v", res)
	}
}
