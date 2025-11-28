//go:build !prod || desktop
// +build !prod desktop

package api

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Singleton instance of the fake OAuth provider
var globalFakeOAuthProvider *FakeOAuthProvider

// getDefaults returns default values for dev builds (same as service package)
func getDefaults() map[string]string {
	slog.Info("authservice:fake_oauth_dev:getDefaults")
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

// getEnvOrDefault returns the environment variable value or the default value (same as service package)
func getEnvOrDefault(key, defaultValue string) string {
	slog.Info("authservice:fake_oauth_dev:getEnvOrDefault", "key", key, "defaultValue", defaultValue)
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Init initializes the auth service API with fake OAuth provider for development
func (a *AuthServiceAPI) Init() {
	slog.Debug("authservice:fake_oauth_dev:Init")
	// Initialize core functionality
	a.initCore()

	// Use singleton pattern to ensure the same provider instance is used throughout
	if globalFakeOAuthProvider == nil {
		globalFakeOAuthProvider = NewFakeOAuthProvider()
		slog.Info("authservice:fake_oauth_dev:Init", "step", "Created singleton fake OAuth provider", "keySize", globalFakeOAuthProvider.privateKey.Size())
	}

	// Register fake OAuth provider under /fakeoauth
	slog.Info("authservice:fake_oauth_dev:Init", "step", "Registered fake OAuth provider under /fakeoauth")
	globalFakeOAuthProvider.RegisterRoutes(a.mux, "/fakeoauth")
}

// FakeOAuthProvider - complete implementation that mimics Google OAuth
type FakeOAuthProvider struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// Use a fixed key pair for development to ensure consistency
// In production, keys would be properly managed and rotated
var devPrivateKey *rsa.PrivateKey
var devPublicKey *rsa.PublicKey

func init() {
	// Generate a fixed key pair for development
	// This ensures the JWKS and token signatures remain consistent
	var err error
	devPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		slog.Error("authservice:fake_oauth_dev:init", "step", "Failed to generate RSA key", "error", err)
		panic("Failed to generate RSA key: " + err.Error())
	}
	devPublicKey = &devPrivateKey.PublicKey

	slog.Info("authservice:fake_oauth_dev:init", "step", "Generated RSA key pair in init()", "keySize", devPrivateKey.Size())

}

func NewFakeOAuthProvider() *FakeOAuthProvider {
	slog.Info("authservice:fake_oauth_dev:NewFakeOAuthProvider")
	return &FakeOAuthProvider{
		privateKey: devPrivateKey,
		publicKey:  devPublicKey,
	}
}

func (f *FakeOAuthProvider) RegisterRoutes(mux *http.ServeMux, basePath string) {
	slog.Info("authservice:fake_oauth_dev:RegisterRoutes", "basePath", basePath)
	mux.HandleFunc(basePath+"/.well-known/openid-configuration", f.oidcDiscoveryHandler)
	mux.HandleFunc(basePath+"/oauth2/v2/auth", f.authHandler)
	mux.HandleFunc(basePath+"/token", f.tokenHandler)
	mux.HandleFunc(basePath+"/.well-known/jwks.json", f.jwksHandler)
	mux.HandleFunc(basePath+"/index.html", f.indexHandler)
	mux.HandleFunc(basePath+"/", f.catchAllHandler)
}

func (f *FakeOAuthProvider) catchAllHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("authservice:fake_oauth_dev:catchAllHandler", "path", r.URL.Path)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("<html></html>"))
}

func (f *FakeOAuthProvider) indexHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("authservice:fake_oauth_dev:indexHandler")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("<html></html>"))
}
func (f *FakeOAuthProvider) authHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("authservice:fake_oauth_dev:authHandler")
	// Simple fake OAuth provider that immediately redirects to callback with a fake code

	// Generate a fake authorization code
	fakeCode := "fake_auth_code_12345"

	// Get the redirect_uri from query params (standard OAuth flow)
	redirectURI := r.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		// Default to local callback if not provided
		redirectURI = "http://localhost:8080/callback"
	}

	// Redirect to callback with fake code
	callbackURL := redirectURI + "?code=" + fakeCode
	// We were send a http redirect, HTTP 302, earlier with statusfound, the browser was able to recognize this
	// and auto redirect us to the new URL
	// but in the wails webview app the 302 was shown as a clickable link "Found", on clicking it goes to the app
	// to fix this issue we are now taking this route of using javascript to redirect

	//http.Redirect(w, r, callbackURL, http.StatusFound)
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	// This injects JS that forces the browser to move
	html := fmt.Sprintf("<script>window.location.href = '%s';</script>", callbackURL)
	w.Write([]byte(html))
}

func (f *FakeOAuthProvider) tokenHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("authservice:fake_oauth_dev:tokenHandler")
	// Simple fake token endpoint that returns a fake ID token

	if r.Method != http.MethodPost {
		slog.Error("authservice:fake_oauth_dev:tokenHandler", "step", "Method not allowed", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		slog.Error("authservice:fake_oauth_dev:tokenHandler", "step", "Invalid form data", "error", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Accept any client credentials (no validation for fake OAuth)
	// In real OAuth, we'd validate client_id and client_secret

	// Check for the fake code
	code := r.FormValue("code")
	if code != "fake_auth_code_12345" {
		slog.Error("authservice:fake_oauth_dev:tokenHandler", "step", "Invalid authorization code", "code", code)
		http.Error(w, "Invalid authorization code", http.StatusBadRequest)
		return
	}

	// Create a fake JWT ID token with proper structure
	// This is still fake but with a more realistic structure that bypasses verification
	fakeIDToken := f.createFakeJWT()

	slog.Info("Token endpoint returning response", "id_token_length", len(fakeIDToken))

	// Return OAuth2 token response
	tokenResponse := map[string]interface{}{
		"access_token":  "fake_access_token",
		"token_type":    "Bearer",
		"expires_in":    3600,
		"id_token":      fakeIDToken,
		"refresh_token": "fake_refresh_token",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokenResponse)
}

func (f *FakeOAuthProvider) oidcDiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("authservice:fake_oauth_dev:oidcDiscoveryHandler")
	// OIDC Discovery document that points to our fake endpoints
	defaults := getDefaults()
	baseURL := getEnvOrDefault("OAUTH_ISSUER_URL", defaults["OAUTH_ISSUER_URL"])

	discovery := map[string]interface{}{
		"issuer":                                baseURL,
		"authorization_endpoint":                baseURL + "/oauth2/v2/auth",
		"token_endpoint":                        baseURL + "/token",
		"userinfo_endpoint":                     baseURL + "/userinfo",
		"jwks_uri":                              baseURL + "/.well-known/jwks.json",
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported":                      []string{"openid", "email", "profile"},
		"claims_supported":                      []string{"sub", "email", "email_verified", "name", "picture"},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(discovery)
}

func (f *FakeOAuthProvider) jwksHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("authservice:fake_oauth_dev:jwksHandler")

	// Convert RSA public key to JWK format with proper base64url encoding
	// Ensure N is correctly encoded (should be big-endian)
	nBytes := f.publicKey.N.Bytes()
	n := base64.RawURLEncoding.EncodeToString(nBytes)

	// E is usually 65537 for RSA keys, encode as big-endian bytes
	eBytes := big.NewInt(int64(f.publicKey.E)).Bytes()
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	slog.Info("authservice:fake_oauth_dev:jwksHandler", "step", "JWKS key details", "n_length", len(nBytes), "e_value", f.publicKey.E)

	jwks := map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "RSA",
				"use": "sig",
				"kid": "fake-key-id",
				"n":   n,
				"e":   e,
				"alg": "RS256",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jwks)
}

func (f *FakeOAuthProvider) createFakeJWT() string {
	slog.Info("authservice:fake_oauth_dev:createFakeJWT")
	// Create JWT payload with current timestamps
	now := time.Now().Unix()
	defaults := getDefaults()
	issuer := getEnvOrDefault("OAUTH_ISSUER_URL", defaults["OAUTH_ISSUER_URL"])
	// Get client ID from environment or use default
	clientID := getEnvOrDefault("GOOGLE_CLIENT_ID", defaults["GOOGLE_CLIENT_ID"])

	slog.Info("authservice:fake_oauth_dev:createFakeJWT", "step", "Creating fake JWT", "issuer", issuer, "clientID", clientID, "now", now, "exp", now+3600)

	payload := map[string]interface{}{
		"iss":            issuer,
		"aud":            clientID, // Must match the client ID used by the verifier
		"sub":            "1234567890",
		"email":          "test@test.com",
		"email_verified": true,
		"name":           "Test User",
		"picture":        "https://example.com/photo.jpg",
		"iat":            now,
		"exp":            now + 3600, // 1 hour from now
	}

	slog.Info("authservice:fake_oauth_dev:createFakeJWT", "step", "JWT payload", "payload", payload)

	// Create JWT using golang-jwt library with RS256 for proper OIDC compliance
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims(payload))

	// Set the key ID in the header to match JWKS
	token.Header["kid"] = "fake-key-id"

	slog.Info("authservice:fake_oauth_dev:createFakeJWT", "step", "JWT header before signing", "header", token.Header)

	tokenString, err := token.SignedString(f.privateKey)
	if err != nil {
		slog.Error("authservice:fake_oauth_dev:createFakeJWT", "step", "Failed to sign JWT with RS256", "error", err)

		// Debug: Try to understand the key
		slog.Error("authservice:fake_oauth_dev:createFakeJWT", "step", "Private key details", "keySize", f.privateKey.Size(), "keyType", fmt.Sprintf("%T", f.privateKey))

		// Fallback to manually created signature for development
		header := map[string]interface{}{
			"alg": "RS256",
			"kid": "fake-key-id",
			"typ": "JWT",
		}

		headerBytes, _ := json.Marshal(header)
		payloadBytes, _ := json.Marshal(payload)

		headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
		payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

		slog.Info("authservice:fake_oauth_dev:createFakeJWT", "step", "Manual JWT creation", "header_b64", headerB64, "payload_b64", payloadB64)

		// Create signature using RSA private key
		signingString := headerB64 + "." + payloadB64
		hash := sha256.Sum256([]byte(signingString))

		signature, signErr := rsa.SignPKCS1v15(rand.Reader, f.privateKey, crypto.SHA256, hash[:])
		if signErr != nil {
			slog.Error("Failed to sign JWT manually", "error", signErr)
			signature = []byte("fake_signature_fallback")
		}

		signatureB64 := base64.RawURLEncoding.EncodeToString(signature)
		tokenString = strings.Join([]string{headerB64, payloadB64, signatureB64}, ".")

		slog.Info("authservice:fake_oauth_dev:createFakeJWT", "step", "Manual JWT created", "signature_b64_length", len(signatureB64))
	}

	slog.Info("authservice:fake_oauth_dev:createFakeJWT", "step", "Final JWT token created", "tokenLength", len(tokenString))

	// Debug: Parse the token back to verify it was created correctly
	parsedToken, parseErr := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			slog.Error("authservice:fake_oauth_dev:createFakeJWT", "step", "Unexpected signing method", "method", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return &f.privateKey.PublicKey, nil
	})

	if parseErr != nil {
		slog.Error("authservice:fake_oauth_dev:createFakeJWT", "step", "Failed to parse created JWT for verification", "error", parseErr)
	} else if parsedToken.Valid {
		slog.Info("authservice:fake_oauth_dev:createFakeJWT", "step", "Successfully created and verified JWT token")
	} else {
		slog.Error("authservice:fake_oauth_dev:createFakeJWT", "step", "Created JWT token is not valid")
	}

	return tokenString
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
