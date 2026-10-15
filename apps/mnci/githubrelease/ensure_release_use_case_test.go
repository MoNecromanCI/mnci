package githubrelease

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type fakeGitHub struct {
	mu       sync.Mutex
	releases map[string]*Release
	uploads  []string
}

func (f *fakeGitHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.Header.Get("Authorization") != "Bearer tok" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	switch {
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/repos/o/r/releases/tags/"):
		rel, ok := f.releases[strings.TrimPrefix(r.URL.Path, "/repos/o/r/releases/tags/")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(rel)
	case r.Method == http.MethodPost && r.URL.Path == "/repos/o/r/releases":
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		rel := &Release{ID: 1, HTMLURL: "https://example/" + body["tag_name"], UploadURL: "http://" + r.Host + "/upload/" + body["tag_name"] + "{?name,label}"}
		f.releases[body["tag_name"]] = rel
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(rel)
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/upload/"):
		raw, _ := io.ReadAll(r.Body)
		name := r.URL.Query().Get("name")
		rel := f.releases[strings.TrimPrefix(r.URL.Path, "/upload/")]
		rel.Assets = append(rel.Assets, Asset{Name: name})
		f.uploads = append(f.uploads, name+"="+string(raw))
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	default:
		w.WriteHeader(http.StatusTeapot)
	}
}

func TestEnsureCreatesOnceAndUploadsOnlyMissingAssets(t *testing.T) {
	fake := &fakeGitHub{releases: map[string]*Release{}}
	srv := httptest.NewServer(fake)
	defer srv.Close()
	c := &Client{Token: "tok", API: srv.URL, HTTP: srv.Client()}
	dir := t.TempDir()
	a, b := filepath.Join(dir, "app_1.0.0_linux_amd64.zip"), filepath.Join(dir, "checksums.txt")
	_ = os.WriteFile(a, []byte("A"), 0o644)
	_ = os.WriteFile(b, []byte("B"), 0o644)
	repo := Repository{Owner: "o", Name: "r"}

	out, err := Ensure(c, repo, Request{Tag: "app@1.0.0", Title: "app 1.0.0", Notes: "n", Assets: []string{a}})
	if err != nil || !out.Created || len(out.Uploaded) != 1 {
		t.Fatalf("first run: %+v %v", out, err)
	}
	out, err = Ensure(c, repo, Request{Tag: "app@1.0.0", Assets: []string{a, b}})
	if err != nil || out.Created || len(out.Uploaded) != 1 || out.Uploaded[0] != "checksums.txt" {
		t.Fatalf("a re-run uploads only what is missing: %+v %v", out, err)
	}
	if len(fake.uploads) != 2 || fake.uploads[0] != "app_1.0.0_linux_amd64.zip=A" {
		t.Fatalf("uploads %v", fake.uploads)
	}
}

func TestEnsureReportsAPIErrors(t *testing.T) {
	srv := httptest.NewServer(&fakeGitHub{releases: map[string]*Release{}})
	defer srv.Close()
	c := &Client{Token: "wrong", API: srv.URL, HTTP: srv.Client()}
	if _, err := Ensure(c, Repository{Owner: "o", Name: "r"}, Request{Tag: "x@1.0.0"}); err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("want a 401 error, got %v", err)
	}
}

func TestNotes(t *testing.T) {
	if got := Notes("dependency x releases", nil); got != "dependency x releases\n" {
		t.Fatal(got)
	}
	if got := Notes("", []string{"feat: a", "fix: b"}); got != "## Changes\n\n- feat: a\n- fix: b\n" {
		t.Fatal(got)
	}
}
