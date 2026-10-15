package scaffolding

var sliceTemplates = map[string]string{
	"[[.Dir]]/greeting/doc.go": `// Package greeting greets people: a worked example of a slice. Replace it with
// your own outcome.
package greeting
`,
	"[[.Dir]]/greeting/greeting_contract.go": `package greeting

// Greeting is a greeting addressed to someone.
type Greeting struct {
	Message string
}
`,
	"[[.Dir]]/greeting/greet_use_case.go": `package greeting

// Greet greets someone by name.
func Greet(name string) Greeting {
	return Greeting{Message: "Hello, " + name + "!"}
}
`,
	"[[.Dir]]/greeting/greet_use_case_test.go": `package greeting

import "testing"

func TestGreet(t *testing.T) {
	if got := Greet("world").Message; got != "Hello, world!" {
		t.Fatalf("got %q", got)
	}
}
`,
}

var appTemplates = map[string]string{
	"[[.Dir]]/main.go": `package main

import (
	"fmt"

	"[[.Module]]/greeting"
)

func main() {
	fmt.Println(greeting.Greet("world").Message)
}
`,
}

var libTemplates = map[string]string{
	"[[.Dir]]/doc.go": `// Package [[.Package]] is the [[.Name]] module. Its code lives in the slice
// packages below this directory, one package per outcome.
package [[.Package]]
`,
}
