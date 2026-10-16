package clicontract

import (
	"sort"
	"strings"

	"github.com/MoNecromanCI/mnci/apps/mnci/pluginhost"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

var kindText = map[protocol.ProjectKind]string{
	protocol.ProjectApp:     "an app: deployed or run, never published",
	protocol.ProjectLib:     "an internal library: imported by other projects, never published",
	protocol.ProjectPackage: "a package: published to its registry on release",
}

// languages names a plugin's projects as the editor's icons know them.
var languages = map[string]string{"node": "typescript"}

// Kinds lists what `mnci add <kind>` accepts: <language>-<app|lib|package>
// for every language plugin that generates projects.
func Kinds(set *pluginhost.PluginSet) []KindDescription {
	var out []KindDescription
	for _, c := range set.Clients {
		if !c.Info.Capabilities.Generate || !pluginhost.HasRole(c, protocol.RoleLanguage) {
			continue
		}
		name := c.Info.Plugin.Name
		lang := languages[name]
		if lang == "" {
			lang = name
		}
		for _, kind := range []protocol.ProjectKind{protocol.ProjectApp, protocol.ProjectLib, protocol.ProjectPackage} {
			out = append(out, KindDescription{
				Kind: name + "-" + string(kind), Language: lang,
				Label:       title(name) + " " + string(kind),
				Description: title(name) + ": " + kindText[kind] + ". Built with the workspace's default stack, or --stack.",
				Flags:       []string{"stack"},
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Kind < out[j].Kind })
	return out
}

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
