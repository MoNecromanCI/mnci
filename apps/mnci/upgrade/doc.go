// Package upgrade is `mnci upgrade`: after installing a newer mnci, bring the
// workspace up to it. mnci.toml is migrated to the current schema, and the
// files mnci owns (the pipelines) are regenerated, a stale one deleted; files
// mnci wrote once (README, .gitignore) and the team's own are left alone.
package upgrade
