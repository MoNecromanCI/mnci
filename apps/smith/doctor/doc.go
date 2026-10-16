// Package doctor checks a workspace's health: smith's own invariants plus
// every plugin's checks. It is read-only; every failed finding names its
// remedy, and the JSON output is a stable contract for editors and CI.
package doctor
