package scaffolding

// templates are every project's: the deployment and the linter rules that
// fail lint (az bicep lint exits 0 on warnings).
var templates = map[string]string{
	"[[.Dir]]/main.bicep": `metadata version = '0.0.0'
metadata smithKind = '[[.Kind]]'
[[- if .Stack]]
metadata smithStack = '[[.Stack]]'
[[- end]]

@description('Where the resources are deployed.')
param location string = resourceGroup().location

@minLength(3)
@maxLength(24)
@description('The storage account name, globally unique.')
param storageName string

resource storage 'Microsoft.Storage/storageAccounts@2023-05-01' = {
  name: storageName
  location: location
  sku: {
    name: 'Standard_LRS'
  }
  kind: 'StorageV2'
}

output storageId string = storage.id
`,
	"[[.Dir]]/bicepconfig.json": `{
  "analyzers": {
    "core": {
      "rules": {
        "no-unused-params": { "level": "error" },
        "no-unused-vars": { "level": "error" },
        "no-hardcoded-location": { "level": "error" }
      }
    }
  }
}
`,
}

// appTemplates are a deployment's parameters; a lib is a module others
// pass them to.
var appTemplates = map[string]string{
	"[[.Dir]]/main.bicepparam": `using './main.bicep'

param storageName = 'st[[.Storage]]'
`,
}
