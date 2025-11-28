//go:build prod && !desktop
// +build prod,!desktop

package main

// getJWTDefaults returns empty defaults for prod builds (must be set via env)
func getJWTDefaults() (string, string) {
	return "", ""
}
