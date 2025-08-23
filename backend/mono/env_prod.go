//go:build prod
// +build prod

package main

// getJWTDefaults returns empty defaults for prod builds (must be set via env)
func getJWTDefaults() (string, string) {
	return "", ""
}
