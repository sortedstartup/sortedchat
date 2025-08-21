package service

import (
	"context"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"sortedstartup/authservice/dao"
	db "sortedstartup/authservice/dao"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type AuthService struct {
	oauthCfg oauth2.Config
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier

	appJWTSecret []byte
	cookieName   string
	cookiePath   string
	tokenTTL     time.Duration
	appIssuer    string

	callbackTemplate *template.Template
	userService      *UserService
}

func NewAuthService(userService *UserService) *AuthService {

	ctx := context.Background()
	// OIDC discovery
	issuer := "https://accounts.google.com"
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	var err error
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		log.Fatalf("oidc provider: %v", err)
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
		log.Fatalf("failed to parse callback template: %v", err)
	}

	return &AuthService{
		oauthCfg:         oauthCfg,
		provider:         provider,
		verifier:         verifier,
		appJWTSecret:     []byte(os.Getenv("APP_JWT_SECRET")),
		cookieName:       "app_jwt",
		cookiePath:       "/",
		tokenTTL:         24 * time.Hour,
		appIssuer:        os.Getenv("APP_ISSUER"),
		callbackTemplate: callbackTemplate,
		userService:      userService,
	}
}

func (s *AuthService) OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	code := q.Get("code")
	// state := q.Get("state")
	if code == "" { //|| state == "" {
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
		http.Error(w, "code exchange failed", http.StatusBadRequest)
		return
	}

	rawIDToken, _ := tok.Extra("id_token").(string)
	if rawIDToken == "" {
		http.Error(w, "no id_token", http.StatusUnauthorized)
		return
	}

	// Verify ID token (signature, iss, aud, exp)
	idToken, err := s.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		http.Error(w, "invalid id_token", http.StatusUnauthorized)
		return
	}
	// Nonce binding check (recommended)
	if idToken.Nonce != "" { // && idToken.Nonce != tmp.nonce {
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
		http.Error(w, "no subject", http.StatusUnauthorized)
		return
	}

	// Upsert user using UserService
	oAuthProvider := "google"
	oAuthUserID := claims.Sub
	roles := "user" // Convert to string for DAO
	isFederated := true

	userID, err := s.userService.CreateUserIfNotExists(claims.Email, roles, oAuthProvider, oAuthUserID, isFederated)
	if err != nil {
		slog.Error("user creation failed", "error", err)
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
		slog.Error("jwt issue failed", "error", err)
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
	// Generate the OAuth URL with proper state parameter for security
	// In a production app, you should generate a random state and store it
	// for validation in the callback
	return s.oauthCfg.AuthCodeURL("state", oauth2.AccessTypeOffline)
}

type UserService struct {
	dao dao.UserDAO
}

func NewUserService(dao dao.UserDAO) *UserService {
	return &UserService{dao: dao}
}

func (u *UserService) Init(config *dao.Config) {
	switch config.Database.Type {
	case db.DatabaseTypeSQLite:
		slog.Info("UserService: Running SQLite migrations")
		if err := db.MigrateSQLite(config.Database.SQLite.URL); err != nil {
			log.Fatalf("UserService: Failed to migrate SQLite database: %v", err)
		}
		if err := db.SeedSqlite(config.Database.SQLite.URL); err != nil {
			log.Fatalf("UserService: Failed to seed SQLite database: %v", err)
		}
	case db.DatabaseTypePostgres:
		slog.Info("UserService: Running PostgreSQL migrations")
		dsn := config.Database.Postgres.GetPostgresDSN()
		if err := db.MigratePostgres(dsn); err != nil {
			log.Fatalf("UserService: Failed to migrate PostgreSQL database: %v", err)
		}
		if err := db.SeedPostgres(dsn); err != nil {
			log.Fatalf("UserService: Failed to seed PostgreSQL database: %v", err)
		}
	default:
		log.Fatalf("UserService: Unsupported database type: %s", config.Database.Type)
	}
}

// GenerateNewUserID generates a new unique user ID using UUID
func GenerateNewUserID() string {
	return uuid.New().String()
}

func (u *UserService) DoesUserExist(userID string) (bool, error) {
	return u.dao.DoesUserExist(userID)
}

func (u *UserService) CreateUserIfNotExists(email, roles, oAuthProvider, oAuthUserID string, isFederated bool) (string, error) {
	// Generate a new user ID
	userID := GenerateNewUserID()

	return u.dao.CreateUserIfNotExists(userID, email, roles, oAuthProvider, oAuthUserID, isFederated)
}
