package transport

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	s := NewStream(&buf, &buf)
	for _, msg := range []string{`{"a":1}`, `{}`, `{"text":"line\nbreak"}`} {
		if err := s.Write([]byte(msg)); err != nil {
			t.Fatal(err)
		}
	}
	for _, want := range []string{`{"a":1}`, `{}`, `{"text":"line\nbreak"}`} {
		got, err := s.Read()
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("got %s want %s", got, want)
		}
	}
	if _, err := s.Read(); !errors.Is(err, io.EOF) {
		t.Fatalf("want EOF, got %v", err)
	}
}

func TestRejectsBadHeaders(t *testing.T) {
	for _, input := range []string{
		"Content-Length: nope\r\n\r\n{}",
		"X-Other: 1\r\n\r\n{}",
		"garbage\r\n\r\n",
		"Content-Length: 10\r\n\r\n{}",
	} {
		s := NewStream(strings.NewReader(input), io.Discard)
		if _, err := s.Read(); err == nil {
			t.Fatalf("expected an error for %q", input)
		}
	}
}

func TestAcceptsLowercaseHeaderAndExtraHeaders(t *testing.T) {
	s := NewStream(strings.NewReader("content-length: 2\r\nContent-Type: application/json\r\n\r\n{}"), io.Discard)
	got, err := s.Read()
	if err != nil || string(got) != "{}" {
		t.Fatalf("got %q, %v", got, err)
	}
}
