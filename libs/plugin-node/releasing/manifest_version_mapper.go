package releasing

import (
	"errors"
	"strconv"
	"strings"
)

var dependencySections = map[string]bool{"dependencies": true, "devDependencies": true, "peerDependencies": true, "optionalDependencies": true}

// stringValue is the byte span of a JSON string value (quotes included) and the
// object keys leading to it.
type stringValue struct {
	path       []string
	start, end int
}

// SetVersions rewrites the top-level "version" and the ranges of the given
// workspace dependencies ("^<version>"), touching only those string values.
// Peer ranges are compatibility statements, never rewritten.
func SetVersions(src []byte, version string, deps map[string]string) ([]byte, error) {
	values, err := stringValues(src)
	if err != nil {
		return nil, err
	}
	out := string(src)
	found := false
	// Replace back to front so earlier offsets stay valid.
	for i := len(values) - 1; i >= 0; i-- {
		v := values[i]
		var replacement string
		switch {
		case len(v.path) == 1 && v.path[0] == "version":
			replacement, found = version, true
		case len(v.path) == 2 && dependencySections[v.path[0]] && v.path[0] != "peerDependencies":
			next, ok := deps[v.path[1]]
			if !ok {
				continue
			}
			replacement = "^" + next
		default:
			continue
		}
		out = out[:v.start] + strconv.Quote(replacement) + out[v.end:]
	}
	if !found {
		return nil, errors.New(`package.json has no top-level "version"`)
	}
	return []byte(out), nil
}

// stringValues scans JSON and returns every string that is an object member's
// value, with its key path. It assumes valid JSON.
func stringValues(src []byte) ([]stringValue, error) {
	var (
		out      []stringValue
		keys     []string // key per open object (empty for arrays)
		inObject []bool
		pending  string
		expectV  bool
	)
	s := string(src)
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '{', '[':
			key := ""
			if expectV {
				key = pending
			}
			keys = append(keys, key)
			inObject = append(inObject, c == '{')
			expectV = false
		case '}', ']':
			if len(keys) == 0 {
				return nil, errors.New("unbalanced JSON")
			}
			keys = keys[:len(keys)-1]
			inObject = inObject[:len(inObject)-1]
			expectV = false
		case ':':
			expectV = true
		case ',':
			expectV = false
		case '"':
			end, text, err := readString(s, i)
			if err != nil {
				return nil, err
			}
			switch {
			case expectV:
				path := append(append([]string{}, nonEmpty(keys[1:])...), pending)
				out = append(out, stringValue{path: path, start: i, end: end})
				expectV = false
			case len(inObject) > 0 && inObject[len(inObject)-1]:
				pending = text
			}
			i = end - 1
		}
	}
	return out, nil
}

func nonEmpty(keys []string) []string {
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		if k != "" {
			out = append(out, k)
		}
	}
	return out
}

func readString(s string, start int) (int, string, error) {
	var b strings.Builder
	for i := start + 1; i < len(s); i++ {
		switch s[i] {
		case '\\':
			if i+1 < len(s) {
				b.WriteByte(s[i+1])
			}
			i++
		case '"':
			return i + 1, b.String(), nil
		default:
			b.WriteByte(s[i])
		}
	}
	return 0, "", errors.New("unterminated string in JSON")
}
