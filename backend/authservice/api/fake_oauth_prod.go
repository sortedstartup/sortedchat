//go:build prod
// +build prod

package api

// Init initializes the auth service API for production (no fake OAuth provider)
func (a *AuthServiceAPI) Init() {
	// Initialize core functionality only
	a.initCore()
}
