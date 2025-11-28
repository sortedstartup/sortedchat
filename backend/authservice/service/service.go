package service

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	dao "sortedstartup/authservice/dao"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type AuthService struct {
	oauthCfg                    oauth2.Config
	oauthProviderURLForFrontend string
	provider                    *oidc.Provider
	verifier                    *oidc.IDTokenVerifier

	appJWTSecret []byte
	cookieName   string
	cookiePath   string
	tokenTTL     time.Duration
	appIssuer    string

	callbackTemplate *template.Template
	userService      *UserService

	// Lazy initialization fields
	initOnce    sync.Once
	initialized bool
	initError   error
}

func NewAuthService(userService *UserService) *AuthService {
	slog.Debug("authservice:service:NewAuthService")
	return &AuthService{
		userService: userService,
		// Lazy initialization - all other fields will be set in initialize()
	}
}

// initialize performs the actual initialization of the AuthService
// This method is called only once when OAuthCallbackHandler is first invoked
func (s *AuthService) initialize() {
	slog.Debug("authservice:service:initialize")
	ctx := context.Background()
	defaults := getDefaults()

	// OIDC discovery - use configurable OAuth provider
	issuer := getEnvOrDefault("OAUTH_ISSUER_URL", defaults["OAUTH_ISSUER_URL"])
	s.oauthProviderURLForFrontend = getEnvOrDefault("OAUTH_PROVIDER_URL_FOR_FRONTEND", defaults["OAUTH_PROVIDER_URL_FOR_FRONTEND"])
	clientID := getEnvOrDefault("GOOGLE_CLIENT_ID", defaults["GOOGLE_CLIENT_ID"])
	clientSecret := getEnvOrDefault("GOOGLE_CLIENT_SECRET", defaults["GOOGLE_CLIENT_SECRET"])
	redirectURL := getEnvOrDefault("GOOGLE_REDIRECT_URL", defaults["GOOGLE_REDIRECT_URL"])

	//using these env variables to start
	slog.Debug("AuthService:service:initialize", "step", "initializing", "issuer", issuer, "clientID", clientID, "clientSecret", "[HIDDEN]", "redirectURL", redirectURL)
	var err error
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		slog.Error("authservice:service:initialize", "step", "oidc provider, issuer", "error", err, "issuer", issuer)
		s.initError = err
		return
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	oauthCfg := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
	}

	// Create the callback HTML template
	callbackHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Authentication Success - SortedChat</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif;
            background: #f9fafb;
            margin: 0;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            text-align: center;
            color: #374151;
        }
        h1 {
            margin-bottom: 0.5rem;
            font-size: 1.5rem;
        }
        p {
            color: #6b7280;
            margin: 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Authentication Successful!</h1>
        <p>Redirecting...</p>
    </div>
    
    <script>
        // Set JWT token in localStorage and redirect immediately
        localStorage.setItem('sortedchat.jwt', '{{.JWT}}');
        console.log('JWT token set in localStorage');
        window.location.href = '/';
    </script>
</body>
</html>`

	callbackTemplate, err := template.New("callback").Parse(callbackHTML)
	if err != nil {
		slog.Error("authservice:service:initialize", "step", "failed to parse callback template", "error", err)
		s.initError = err
		return
	}

	// Set all the initialized fields
	s.oauthCfg = oauthCfg
	s.provider = provider
	s.verifier = verifier
	s.appJWTSecret = []byte(getEnvOrDefault("APP_JWT_SECRET", defaults["APP_JWT_SECRET"]))
	s.cookieName = "app_jwt"
	s.cookiePath = "/"
	s.tokenTTL = 24 * time.Hour
	s.appIssuer = getEnvOrDefault("APP_ISSUER", defaults["APP_ISSUER"])
	s.callbackTemplate = callbackTemplate
	s.initialized = true

	// Debug: Log the JWT configuration being used
	slog.Info("AuthService JWT configuration",
		"appIssuer", s.appIssuer,
		"secretLength", len(s.appJWTSecret),
		"secretPrefix", string(s.appJWTSecret[:min(10, len(s.appJWTSecret))]))
}

func (s *AuthService) OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("authservice:service:OAuthCallbackHandler")
	// Lazy initialization - initialize only on first call
	s.initOnce.Do(s.initialize)

	// Check if initialization failed
	if s.initError != nil {
		slog.Error("authservice:service:OAuthCallbackHandler", "step", "AuthService initialization failed", "error", s.initError)
		http.Error(w, "Service initialization failed", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	q := r.URL.Query()

	code := q.Get("code")
	// state := q.Get("state")
	if code == "" { //|| state == "" {
		slog.Debug("missing code/state", "code", code)
		http.Error(w, "missing code/state", http.StatusBadRequest)
		return
	}

	// tmp, ok := tmpStore.take(state) // validates state once
	// if !ok || time.Now().After(tmp.exp) {
	// 	http.Error(w, "invalid state", http.StatusBadRequest)
	// 	return
	// }

	// Exchange code + PKCE
	tok, err := s.oauthCfg.Exchange(ctx, code)
	//oauth2.SetAuthURLParam("code_verifier", tmp.codeVerifier))
	if err != nil {
		slog.Error("authservice:service:OAuthCallbackHandler", "step", "code exchange failed", "error", err)
		http.Error(w, "code exchange failed", http.StatusBadRequest)
		return
	}

	rawIDToken, _ := tok.Extra("id_token").(string)
	if rawIDToken == "" {
		slog.Error("authservice:service:OAuthCallbackHandler", "step", "no id_token", "rawIDToken", rawIDToken)
		http.Error(w, "no id_token", http.StatusUnauthorized)
		return
	}

	// Debug the raw ID token before verification
	slog.Debug("Attempting to verify ID token", "rawIDToken_length", len(rawIDToken), "rawIDToken_prefix", rawIDToken[:min(100, len(rawIDToken))])

	// Verify ID token (signature, iss, aud, exp)
	idToken, err := s.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		slog.Error("authservice:service:OAuthCallbackHandler", "step", "id token verification failed", "error", err, "rawIDToken_length", len(rawIDToken))

		// Try to parse the token without verification to see its contents
		if debugToken, parseErr := jwt.Parse(rawIDToken, func(token *jwt.Token) (interface{}, error) {
			return nil, fmt.Errorf("debug parse - not verifying")
		}); parseErr == nil || debugToken != nil {
			slog.Info("authservice:service:OAuthCallbackHandler", "step", "Debug token info", "header", debugToken.Header, "claims", debugToken.Claims)
		} else {
			slog.Error("authservice:service:OAuthCallbackHandler", "step", "Could not even parse token for debugging", "parseError", parseErr)
		}

		http.Error(w, "invalid id_token", http.StatusUnauthorized)
		return
	}

	// Success case - log it!
	slog.Debug("ID token verification SUCCEEDED!", "subject", idToken.Subject, "issuer", idToken.Issuer)
	// Nonce binding check (recommended)
	if idToken.Nonce != "" { // && idToken.Nonce != tmp.nonce {
		slog.Error("authservice:service:OAuthCallbackHandler", "step", "nonce mismatch", "nonce", idToken.Nonce)
		http.Error(w, "nonce mismatch", http.StatusUnauthorized)
		return
	}

	// Extract identity claims
	var claims struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	_ = idToken.Claims(&claims)
	if claims.Sub == "" {
		slog.Error("authservice:service:OAuthCallbackHandler", "step", "no subject", "claims", claims)
		http.Error(w, "no subject", http.StatusUnauthorized)
		return
	}

	// Upsert user using UserService
	oAuthProvider := "google"
	oAuthUserID := claims.Sub
	roles := "user" // Convert to string for DAO
	isFederated := true
	email := claims.Email

	// Check if email is in the allowlist
	if !isEmailAllowed(email) {
		slog.Debug("authservice:service:OAuthCallbackHandler", "step", "Login attempt from unauthorized email", "email", email)
		http.Error(w, "Access denied: Your email is not authorized to access this application", http.StatusForbidden)
		return
	}

	userID, err := s.userService.CreateUserIfNotExists(email, roles, oAuthProvider, oAuthUserID, isFederated)
	if err != nil {
		slog.Error("authservice:service:OAuthCallbackHandler", "step", "user creation failed", "error", err)
		http.Error(w, "user creation failed", http.StatusInternalServerError)
		return
	}

	// Mint your app JWT
	now := time.Now()
	appJWT, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":   s.appIssuer,
		"sub":   userID,
		"email": claims.Email,
		"roles": []string{roles}, // Convert back to array for JWT
		"iat":   now.Unix(),
		"exp":   now.Add(s.tokenTTL).Unix(),
	}).SignedString(s.appJWTSecret)
	if err != nil {
		slog.Error("authservice:service:OAuthCallbackHandler", "step", "jwt issue failed", "error", err)
		http.Error(w, "jwt issue failed", http.StatusInternalServerError)
		return
	}

	// Return HTML page with JWT embedded in JavaScript
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := struct {
		JWT string
	}{
		JWT: appJWT,
	}

	if err := s.callbackTemplate.Execute(w, data); err != nil {
		http.Error(w, "template execution failed", http.StatusInternalServerError)
		return
	}
}

func (s *AuthService) GetAuthURL() string {
	slog.Info("authservice:service:GetAuthURL")
	// Lazy initialization - initialize only on first call
	s.initOnce.Do(s.initialize)

	// Check if initialization failed
	if s.initError != nil {
		slog.Error("authservice:service:GetAuthURL", "step", "AuthService initialization failed during GetAuthURL", "error", s.initError)
		return ""
	}

	// Generate the OAuth URL with proper state parameter for security
	// In a production app, you should generate a random state and store it
	// for validation in the callback
	return s.oauthCfg.AuthCodeURL("state", oauth2.AccessTypeOffline)
}

type OAuthFrontendConfig struct {
	ClientID         string
	OAuthProviderURL string
	OAuthRedirectURL string
}

func (s *AuthService) GetOAuthConfigForFrontend() OAuthFrontendConfig {
	slog.Info("GetOAuthConfigForFrontend")
	return OAuthFrontendConfig{
		ClientID:         s.oauthCfg.ClientID,
		OAuthProviderURL: s.oauthProviderURLForFrontend,
		OAuthRedirectURL: s.oauthCfg.RedirectURL,
	}
}

type UserService struct {
	dao dao.UserDAO
}

func NewUserService(dao dao.UserDAO) *UserService {
	slog.Debug("authservice:service:NewUserService")
	return &UserService{dao: dao}
}

func (u *UserService) Init(config *dao.Config) {
	slog.Debug("authservice:service:Init")
	switch config.Database.Type {
	case dao.DatabaseTypeSQLite:
		slog.Info("UserService: Running SQLite migrations")
		if err := dao.MigrateSQLite(config.Database.SQLite.URL); err != nil {
			log.Fatalf("UserService: Failed to migrate SQLite database: %v", err)
		}
		if err := dao.SeedSqlite(config.Database.SQLite.URL); err != nil {
			log.Fatalf("UserService: Failed to seed SQLite database: %v", err)
		}
	case dao.DatabaseTypePostgres:
		slog.Info("UserService: Running PostgreSQL migrations")
		dsn := config.Database.Postgres.GetPostgresDSN()
		if err := dao.MigratePostgres(dsn); err != nil {
			log.Fatalf("UserService: Failed to migrate PostgreSQL database: %v", err)
		}
		if err := dao.SeedPostgres(dsn); err != nil {
			log.Fatalf("UserService: Failed to seed PostgreSQL database: %v", err)
		}
	default:
		log.Fatalf("UserService: Unsupported database type: %s", config.Database.Type)
	}
}

// GenerateNewUserID generates a new unique user ID using UUID
func GenerateNewUserID() string {
	slog.Info("authservice:service:GenerateNewUserID")
	return uuid.New().String()
}

func (u *UserService) DoesUserExist(userID string) (bool, error) {
	slog.Info("authservice:service:DoesUserExist", "userID", userID)
	return u.dao.DoesUserExist(userID)
}

func (u *UserService) CreateUserIfNotExists(email, roles, oAuthProvider, oAuthUserID string, isFederated bool) (string, error) {
	slog.Info("authservice:service:CreateUserIfNotExists", "email", email, "roles", roles, "oAuthProvider", oAuthProvider, "oAuthUserID", oAuthUserID, "isFederated", isFederated)
	// Generate a new user ID
	userID := GenerateNewUserID()

	return u.dao.CreateUserIfNotExists(userID, email, roles, oAuthProvider, oAuthUserID, isFederated)
}

// getEnvOrDefault returns the environment variable value or the default value
func getEnvOrDefault(key, defaultValue string) string {
	slog.Info("authservice:service:getEnvOrDefault", "key", key, "defaultValue", defaultValue)
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isEmailAllowed checks if the given email is in the allowlist from ALLOWED_LOGIN_EMAILS env variable
// The environment variable should contain comma-separated email addresses
// If ALLOWED_LOGIN_EMAILS is not set or empty, all emails are allowed
func isEmailAllowed(email string) bool {
	slog.Info("authservice:service:isEmailAllowed", "email", email)
	allowedEmails := os.Getenv("ALLOWED_LOGIN_EMAILS")

	// If no allowlist is configured, allow all emails
	if allowedEmails == "" {
		return true
	}

	// Split the comma-separated list and check each email
	emails := strings.Split(allowedEmails, ",")
	for _, allowedEmail := range emails {
		// Trim whitespace and compare (case-insensitive)
		if strings.TrimSpace(strings.ToLower(allowedEmail)) == strings.ToLower(email) {
			return true
		}
	}

	return false
}
