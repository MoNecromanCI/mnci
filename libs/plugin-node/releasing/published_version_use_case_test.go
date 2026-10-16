package releasing

import (
	"errors"
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestPublished(t *testing.T) {
	p := protocol.PublishedParams{Project: protocol.Project{Exports: []string{"@acme/sdk"}}}
	for _, c := range []struct {
		out, errOut string
		err         error
		want        protocol.PublishedResult
	}{
		{"1.4.0\n", "", nil, protocol.PublishedResult{Version: "1.4.0"}},
		{"", "npm error code E404\nnpm error 404 Not Found - GET https://registry/@acme%2fsdk", errors.New("exit 1"), protocol.PublishedResult{}},
		{"", "npm error code ECONNREFUSED\nnpm error connect ECONNREFUSED", errors.New("exit 1"), protocol.PublishedResult{Unreachable: true, Detail: "npm error code ECONNREFUSED"}},
	} {
		capture = func(_ string, argv []string) (string, string, error) {
			if argv[1] == "config" {
				return "https://registry.example/\n", "", nil
			}
			return c.out, c.errOut, c.err
		}
		got, err := Published(t.TempDir(), p)
		c.want.Registry = "https://registry.example/"
		if err != nil || got != c.want {
			t.Errorf("got %+v %v, want %+v", got, err, c.want)
		}
	}
}
