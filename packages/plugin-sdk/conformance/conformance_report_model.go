package conformance

import (
	"fmt"
	"strings"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// Result is one check.
type Result struct {
	Check  string `json:"check"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}

// Report is every check run against one plugin.
type Report struct {
	Plugin  protocol.Peer `json:"plugin"`
	Results []Result      `json:"results"`
}

// OK reports whether every check passed.
func (r Report) OK() bool {
	for _, res := range r.Results {
		if !res.OK {
			return false
		}
	}
	return true
}

// Failures lists the failed checks, one per line.
func (r Report) Failures() string {
	var lines []string
	for _, res := range r.Results {
		if !res.OK {
			lines = append(lines, fmt.Sprintf("%s: %s", res.Check, res.Detail))
		}
	}
	return strings.Join(lines, "\n")
}

func (r *Report) add(check string, ok bool, detail string, args ...any) {
	if ok {
		detail = ""
	} else {
		detail = fmt.Sprintf(detail, args...)
	}
	r.Results = append(r.Results, Result{Check: check, OK: ok, Detail: detail})
}
