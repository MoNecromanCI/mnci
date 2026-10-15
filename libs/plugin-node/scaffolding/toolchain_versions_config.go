package scaffolding

// NodeVersion is the Node.js major the generated pipeline installs.
const NodeVersion = "24"

// devDependencies are the root tooling versions a new workspace pins. Nothing
// installs "latest"; bump these deliberately. TypeScript stays on 6.x until
// ts-jest and typescript-eslint support 7.
var devDependencies = map[string]string{
	"@mnci/eslint-config": "^0.3.31",
	"@types/jest":         "^30.0.0",
	"@types/node":         "^24.19.2",
	"eslint":              "^10.12.0",
	"jest":                "^30.5.2",
	"ts-jest":             "^29.4.15",
	"typescript":          "~6.0.3",
}
