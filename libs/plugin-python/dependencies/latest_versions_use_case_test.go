package dependencies

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestLatest(t *testing.T) {
	realNewest, realFetch := Newest, FetchJSON
	defer func() { Newest, FetchJSON = realNewest, realFetch }()
	t.Setenv("PIP_INDEX_URL", "https://pypi.example/simple")
	var asked string
	Newest = func(index, name string) (string, error) {
		asked = index
		switch name {
		case "requests":
			return "2.32.3", nil
		case "offline":
			return "", errors.New("timeout")
		}
		return "", nil
	}
	FetchJSON = func(_ string, v any) error {
		return json.Unmarshal([]byte(`[{"cycle":"3.14"},{"cycle":"3.13"}]`), v)
	}
	res, _ := Latest("/ws", protocol.LatestParams{Names: []string{"requests", "offline", "nope"}, Tools: true})
	if asked != "https://pypi.example/simple/" {
		t.Errorf("index %q", asked)
	}
	if res.Latest["requests"] != "2.32.3" || res.Unreachable["offline"] != "timeout" || len(res.Latest) != 1 {
		t.Fatalf("%+v", res)
	}
	if len(res.Tools) != 1 || res.Tools[0].Latest != "3.14" || !res.Tools[0].Newer {
		t.Fatalf("tools %+v", res.Tools)
	}
}
