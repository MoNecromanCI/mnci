package pluginhost

import (
	"fmt"
	"slices"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ValidateStack checks a stack against the loaded plugins: every plugin it
// names is loaded and can play the role it is listed under, applies to the
// stack's language, and no two of its plugins provide the same task.
func ValidateStack(st protocol.Stack, set *PluginSet) error {
	lang := set.Get(st.Language)
	if lang == nil {
		return fmt.Errorf("language plugin %q is not loaded", st.Language)
	}
	if !HasRole(lang, protocol.RoleLanguage) {
		return fmt.Errorf("%s is not a language plugin", st.Language)
	}
	type member struct {
		name string
		role protocol.Role
	}
	var members []member
	for _, n := range st.Toolchain {
		members = append(members, member{n, protocol.RoleToolchain})
	}
	for _, n := range st.Structure {
		members = append(members, member{n, protocol.RoleStructure})
	}
	if st.Scaffold != "" {
		members = append(members, member{st.Scaffold, protocol.RoleScaffold})
	}
	for _, n := range st.Deploy {
		members = append(members, member{n, protocol.RoleDeploy})
	}
	seen := map[member]bool{}
	checked := map[string]bool{}
	providedBy := map[string]string{}
	for _, m := range members {
		// One plugin may play several roles in a stack (typescript is the
		// scaffold and a toolchain), but each role lists it once.
		if seen[m] {
			return fmt.Errorf("%s is listed twice as %s", m.name, m.role)
		}
		seen[m] = true
		c := set.Get(m.name)
		if c == nil {
			return fmt.Errorf("plugin %q is not loaded", m.name)
		}
		if !HasRole(c, m.role) {
			return fmt.Errorf("%s is listed as %s, but it plays %v", m.name, m.role, c.Info.EffectiveRoles())
		}
		if len(c.Info.Languages) > 0 && !overlaps(c.Info.Languages, append([]string{lang.Info.Plugin.Name}, lang.Info.Languages...)) {
			return fmt.Errorf("%s applies to %v, not to %s projects", m.name, c.Info.Languages, st.Language)
		}
		if m.role == protocol.RoleScaffold && !c.Supports(protocol.MethodGenerate) {
			return fmt.Errorf("scaffold %s cannot generate projects", m.name)
		}
		if checked[m.name] {
			continue
		}
		checked[m.name] = true
		for _, task := range c.Info.Provides {
			if other, dup := providedBy[task]; dup {
				return fmt.Errorf("%s and %s both provide the %s task; keep one", other, m.name, task)
			}
			providedBy[task] = m.name
		}
	}
	return nil
}

func overlaps(a, b []string) bool {
	for _, x := range a {
		if slices.Contains(b, x) {
			return true
		}
	}
	return false
}
