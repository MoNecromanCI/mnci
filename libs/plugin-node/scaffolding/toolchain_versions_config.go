package scaffolding

// NodeVersion is the Node.js major the generated pipeline installs.
const NodeVersion = "24"

// devDependencies are the root tooling versions a new workspace pins. Nothing
// installs "latest"; bump these deliberately. Lint and test tooling comes
// from conventions (eslint, jest, vitest). TypeScript stays on 6.x until
// ts-jest and typescript-eslint support 7.
var devDependencies = map[string]string{
	"@types/node": "^24.19.2",
	"typescript":  "~6.0.3",
}
