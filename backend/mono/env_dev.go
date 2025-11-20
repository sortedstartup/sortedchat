//go:build !prod || desktop
// +build !prod desktop

package main

// getJWTDefaults returns default JWT values for dev builds
func getJWTDefaults() (string, string) {
	return "fake_jwt_secret_for_dev_only", "sortedchat-dev"
}
