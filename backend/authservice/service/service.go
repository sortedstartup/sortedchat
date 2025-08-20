package service

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
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
}

func NewAuthService() *AuthService {

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
<html>
<head>
    <title>Authentication Success</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background-color: #f5f5f5;
        }
        .container {
            text-align: center;
            background: white;
            padding: 2rem;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .button {
            background-color: #007bff;
            color: white;
            padding: 12px 24px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            text-decoration: none;
            display: inline-block;
        }
        .button:hover {
            background-color: #0056b3;
        }
    </style>
</head>
<body>
    <div class="container">
        <h2>Authentication Successful!</h2>
        <p>You have been successfully authenticated.</p>
        <button class="button" onclick="continueToApp()">Continue to Main App</button>
    </div>
    
    <script>
        function continueToApp() {
            // Set JWT token in localStorage
            localStorage.setItem('sortedchat.jwt', '{{.JWT}}');
            console.log('JWT token set in localStorage:', '{{.JWT}}');
            
            // Wait for 5 seconds, then redirect to main app
            console.log('Waiting 5 seconds before redirecting...');
            setTimeout(function() {
                console.log('Redirecting to main app...');
                window.location.href = '/';
            }, 5000);
        }
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

	// Upsert user (stub). In real app: look up by Sub, create if missing.
	userID := claims.Sub
	roles := []string{"user"}

	// Mint your app JWT
	now := time.Now()
	fmt.Println("issuer", s.appIssuer)
	appJWT, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":   s.appIssuer,
		"sub":   userID,
		"email": claims.Email,
		"roles": roles,
		"iat":   now.Unix(),
		"exp":   now.Add(s.tokenTTL).Unix(),
	}).SignedString(s.appJWTSecret)
	if err != nil {
		http.Error(w, "jwt issue failed", http.StatusInternalServerError)
		return
	}

	fmt.Println("Issuing new app jwt: ", appJWT)

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
