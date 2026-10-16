package convertplan

import (
	"fmt"
	"strings"
)

// RewriteScript turns a root package.json script that drives Nx (or the old
// mnci CLI) into the mnci command doing the same. names maps Nx project
// names to mnci names. ok is false when part of it has no mnci equivalent.
func RewriteScript(script string, names map[string]string) (string, bool) {
	parts := strings.Split(script, "&&")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		words := strings.Fields(part)
		if len(words) > 0 && words[0] == "npx" {
			words = words[1:]
		}
		if len(words) == 0 {
			return "", false
		}
		if words[0] != "nx" && words[0] != "mnci" {
			out = append(out, strings.TrimSpace(part))
			continue
		}
		rewritten, ok := rewriteOne(words, names)
		if !ok {
			return "", false
		}
		out = append(out, rewritten)
	}
	return strings.Join(out, " && "), true
}

func rewriteOne(words []string, names map[string]string) (string, bool) {
	tool, args := words[0], words[1:]
	if len(args) == 0 {
		return "", false
	}
	if tool == "mnci" {
		// The old CLI's `mnci lint [--affected] [--base x]`.
		if args[0] == "lint" {
			return "mnci run lint" + flags(args[1:]), true
		}
		return "", false
	}
	switch args[0] {
	case "run-many", "affected":
		targets, rest, ok := targetsOf(args[1:])
		if !ok {
			return "", false
		}
		cmd := "mnci run " + strings.Join(targets, " ")
		if args[0] == "affected" {
			cmd += " --affected"
		}
		for _, r := range rest {
			name, value, _ := strings.Cut(r, "=")
			switch name {
			case "--projects", "-p":
				var projects []string
				for _, p := range strings.Split(value, ",") {
					if n, ok := names[p]; ok {
						projects = append(projects, n)
					} else {
						return "", false
					}
				}
				cmd += " -p " + strings.Join(projects, ",")
			case "--base":
				cmd += " --base " + value
			case "--parallel", "--output-style", "--skip-nx-cache", "--nxBail", "--verbose":
			default:
				return "", false
			}
		}
		return cmd, true
	case "run":
		if len(args) != 2 {
			return "", false
		}
		project, target, ok := strings.Cut(args[1], ":")
		target, _, _ = strings.Cut(target, ":")
		if n, known := names[project]; ok && known {
			return fmt.Sprintf("mnci run %s -p %s", target, n), true
		}
		return "", false
	case "graph":
		return "mnci graph", true
	case "show":
		if len(args) > 1 && args[1] == "projects" {
			return "mnci projects", true
		}
	case "release":
		for _, a := range args[1:] {
			if a != "--dry-run" && a != "-d" {
				return "", false
			}
		}
		return "mnci release" + flags(args[1:]), true
	}
	if len(args) == 2 {
		// `nx build api` is `nx run api:build`.
		if n, known := names[args[1]]; known {
			return fmt.Sprintf("mnci run %s -p %s", args[0], n), true
		}
	}
	return "", false
}

// targetsOf reads -t/--targets/--target values (comma or space separated)
// and returns the remaining flags.
func targetsOf(args []string) (targets, rest []string, ok bool) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		name, value, hasValue := strings.Cut(a, "=")
		switch name {
		case "-t", "--targets", "--target":
			if !hasValue {
				for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					i++
					targets = append(targets, strings.Split(args[i], ",")...)
				}
				continue
			}
			targets = append(targets, strings.Split(value, ",")...)
		default:
			if strings.HasPrefix(a, "-") && !hasValue && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				a += "=" + args[i+1]
				i++
			}
			rest = append(rest, a)
		}
	}
	return targets, rest, len(targets) > 0
}

func flags(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return " " + strings.Join(args, " ")
}
