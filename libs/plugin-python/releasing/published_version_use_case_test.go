package releasing

import (
	"errors"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestPublished(t *testing.T) {
	t.Setenv("TWINE_REPOSITORY_URL", "http://localhost:8080/")
	t.Setenv("SMITH_PYPI_SIMPLE_URL", "")
	p := protocol.PublishedParams{Project: protocol.Project{Name: "pykit", Exports: []string{"acme_pykit"}}}
	page := `<html><body>
<a href="/packages/acme_pykit-0.9.0.tar.gz#sha256=1">acme_pykit-0.9.0.tar.gz</a>
<a href="/packages/acme_pykit-0.10.1-py3-none-any.whl#sha256=2">acme_pykit-0.10.1-py3-none-any.whl</a>
<a href="/packages/acme_pykit-0.10.0.tar.gz">acme_pykit-0.10.0.tar.gz</a>
</body></html>`
	var asked string
	httpGet = func(url string) (string, error) { asked = url; return page, nil }
	got, _ := Published(p)
	if got.Version != "0.10.1" || asked != "http://localhost:8080/simple/acme-pykit/" {
		t.Fatalf("got %+v from %s", got, asked)
	}
	httpGet = func(string) (string, error) { return "", nil }
	if got, _ := Published(p); got.Version != "" || got.Unreachable {
		t.Fatalf("an unknown project is not published: %+v", got)
	}
	httpGet = func(string) (string, error) { return "", errors.New("dial tcp: connection refused") }
	if got, _ := Published(p); !got.Unreachable || got.Detail == "" {
		t.Fatalf("an unreachable index says so: %+v", got)
	}
}
