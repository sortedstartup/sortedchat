//go:build !prod || desktop
// +build !prod desktop

package service

import "log/slog"

// getDefaults returns default values for dev builds
func getDefaults() map[string]string {
	slog.Info("authservice:service:getDefaults")
	return map[string]string{
		"OAUTH_ISSUER_URL":                "http://localhost:8080/fakeoauth",
		"OAUTH_PROVIDER_URL_FOR_FRONTEND": "http://localhost:5173/hack/fakeoauth/oauth2/v2/auth",
		"GOOGLE_CLIENT_ID":                "fake_client_id",
		"GOOGLE_CLIENT_SECRET":            "fake_client_secret",
		"GOOGLE_REDIRECT_URL":             "http://localhost:5173/hack/callback",
		"APP_JWT_SECRET":                  "fake_jwt_secret_for_dev_only",
		"APP_ISSUER":                      "sortedchat-dev",
	}
}
