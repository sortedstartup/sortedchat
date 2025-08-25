//go:build prod
// +build prod

package service

// getDefaults returns empty defaults for prod builds (must be set via env)
func getDefaults() map[string]string {
	return map[string]string{
		"OAUTH_ISSUER_URL":                "",
		"OAUTH_PROVIDER_URL_FOR_FRONTEND": "http://localhost:5173/hack/fakeoauth/oauth2/v2/auth",
		"GOOGLE_CLIENT_ID":                "",
		"GOOGLE_CLIENT_SECRET":            "",
		"GOOGLE_REDIRECT_URL":             "",
		"APP_JWT_SECRET":                  "",
		"APP_ISSUER":                      "",
	}
}
