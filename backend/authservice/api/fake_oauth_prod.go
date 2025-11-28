//go:build prod && !desktop
// +build prod,!desktop

package api

import "log/slog"

// Init initializes the auth service API for production (no fake OAuth provider)
func (a *AuthServiceAPI) Init() {
	slog.Info("authservice:fake_oauth_prod:Init")
	// Initialize core functionality only
	a.initCore()
}
