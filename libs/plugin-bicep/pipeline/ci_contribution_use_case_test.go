package pipeline

import (
	"errors"
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestCI(t *testing.T) {
	if res := CI(protocol.CIParams{}); len(res.Toolchains) != 0 {
		t.Error("no projects, no toolchain")
	}
	if res := CI(protocol.CIParams{Projects: []protocol.Project{{}}}); len(res.Toolchains) != 1 || res.Toolchains[0].Install == "" {
		t.Errorf("%+v", res)
	}
}

func TestDoctor(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	probe = func(string, []string) (string, error) { return "", errors.New("missing") }
	res := Doctor(t.TempDir(), protocol.DoctorParams{Projects: []protocol.Project{{}}})
	if len(res.Findings) != 1 || res.Findings[0].OK {
		t.Fatalf("without az, only az is reported: %+v", res)
	}
	probe = func(string, []string) (string, error) { return "", nil }
	if res := Doctor(t.TempDir(), protocol.DoctorParams{Projects: []protocol.Project{{}}}); len(res.Findings) != 2 || !res.Findings[1].OK {
		t.Fatalf("%+v", res)
	}
}
