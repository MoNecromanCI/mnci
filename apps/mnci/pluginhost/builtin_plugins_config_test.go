package pluginhost

import (
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/conformance"
)

// Every built-in plugin passes the same conformance suite an external one
// is held to (generation is covered by the end-to-end tests).
func TestBuiltinsConform(t *testing.T) {
	for _, name := range append(BuiltinNames(), sortedNames(BuiltinAddons)...) {
		ctor, _ := Builtin(name)
		report := conformance.Check(connect(t, ctor("1.0.0")), conformance.Options{})
		if !report.OK() {
			t.Errorf("%s:\n%s", name, report.Failures())
		}
	}
}
