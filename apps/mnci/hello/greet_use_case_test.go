package hello

import "testing"

func TestGreet(t *testing.T) {
	if got := Greet("world").Message; got != "Hello, world!" {
		t.Fatalf("got %q", got)
	}
}
