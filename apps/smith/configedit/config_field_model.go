package configedit

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/MoNecromanCI/mnci/apps/mnci/workspace"
)

// Field is one setting of a fixed table of mnci.toml, as the schema
// describes it.
type Field struct {
	// Path is "table.key", e.g. "ci.devcontainer".
	Path        string   `json:"path"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
	Minimum     *int     `json:"minimum,omitempty"`
	MinLength   int      `json:"minLength,omitempty"`
}

// Table and Key split the path.
func (f Field) Table() string { return strings.SplitN(f.Path, ".", 2)[0] }

// Key is the setting's name inside its table.
func (f Field) Key() string { return strings.SplitN(f.Path, ".", 2)[1] }

type schemaNode struct {
	Type                 string                `json:"type"`
	Description          string                `json:"description"`
	Properties           map[string]schemaNode `json:"properties"`
	Items                *schemaNode           `json:"items"`
	Enum                 []string              `json:"enum"`
	Minimum              *int                  `json:"minimum"`
	MinLength            int                   `json:"minLength"`
	AdditionalProperties json.RawMessage       `json:"additionalProperties"`
}

// Fields are the settings of mnci.toml's fixed tables ([workspace], [ci],
// [release], [cache], [plugins]), in schema order of tables and by name
// within one. Tables of named entries (stacks, tasks, projects, ...) are
// edited in the file.
func Fields() ([]Field, error) {
	var root schemaNode
	if err := json.Unmarshal(workspace.Schema, &root); err != nil {
		return nil, err
	}
	var out []Field
	for _, table := range []string{"workspace", "ci", "release", "cache", "plugins"} {
		node := root.Properties[table]
		keys := make([]string, 0, len(node.Properties))
		for k := range node.Properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			p := node.Properties[k]
			if p.Type == "object" {
				continue
			}
			f := Field{Path: table + "." + k, Type: p.Type, Description: p.Description, Minimum: p.Minimum, MinLength: p.MinLength}
			if p.Items != nil {
				f.Enum = p.Items.Enum
			}
			out = append(out, f)
		}
	}
	return out, nil
}

// Find returns the field at path.
func Find(path string) (Field, error) {
	fields, err := Fields()
	if err != nil {
		return Field{}, err
	}
	for _, f := range fields {
		if f.Path == path {
			return f, nil
		}
	}
	var paths []string
	for _, f := range fields {
		paths = append(paths, f.Path)
	}
	return Field{}, fmt.Errorf("%s is not a setting mnci config edits; it edits %s", path, strings.Join(paths, ", "))
}

// Encode checks args against the field and renders them as a TOML value: one
// arg for a scalar, any number for an array.
func (f Field) Encode(args []string) (string, error) {
	if f.Type != "array" && len(args) != 1 {
		return "", fmt.Errorf("%s takes one value", f.Path)
	}
	switch f.Type {
	case "boolean":
		b, err := strconv.ParseBool(args[0])
		if err != nil {
			return "", fmt.Errorf("%s is true or false, not %q", f.Path, args[0])
		}
		return strconv.FormatBool(b), nil
	case "integer":
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return "", fmt.Errorf("%s is a whole number, not %q", f.Path, args[0])
		}
		if f.Minimum != nil && n < *f.Minimum {
			return "", fmt.Errorf("%s is at least %d", f.Path, *f.Minimum)
		}
		return strconv.Itoa(n), nil
	case "string":
		if len(args[0]) < f.MinLength {
			return "", fmt.Errorf("%s cannot be empty", f.Path)
		}
		return Quote(args[0]), nil
	case "array":
		quoted := make([]string, 0, len(args))
		var seen []string
		for _, a := range args {
			if len(f.Enum) > 0 && !slices.Contains(f.Enum, a) {
				return "", fmt.Errorf("%s takes %s, not %q", f.Path, strings.Join(f.Enum, " or "), a)
			}
			if len(f.Enum) > 0 && slices.Contains(seen, a) {
				return "", fmt.Errorf("%s lists %s twice", f.Path, a)
			}
			seen = append(seen, a)
			quoted = append(quoted, Quote(a))
		}
		return "[" + strings.Join(quoted, ", ") + "]", nil
	}
	return "", fmt.Errorf("%s has a type mnci config does not edit (%s)", f.Path, f.Type)
}

// Quote renders a TOML basic string.
func Quote(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\t", `\t`, "\r", `\r`)
	return `"` + r.Replace(s) + `"`
}
