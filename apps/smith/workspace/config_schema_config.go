package workspace

import _ "embed"

// Schema is the JSON Schema of mnci.toml, for editors (Even Better TOML reads
// the #:schema line mnci new writes) and `mnci schema`.
//
//go:embed smith.schema.json
var Schema []byte

// SchemaURL is where the schema is served; it is the file in this repository.
const SchemaURL = "https://raw.githubusercontent.com/MoNecromanCI/mnci/main/apps/mnci/workspace/mnci.schema.json"
