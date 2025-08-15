package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

var (
	oidcProvider *oidc.Provider
	oauth2Config oauth2.Config
	state        string
)

func init() {
	// Load environment variables
	godotenv.Load()

	// Generate random state
	b := make([]byte, 32)
	rand.Read(b)
	state = base64.URLEncoding.EncodeToString(b)
}

func main() {
	// Initialize OIDC provider
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		log.Fatal("Failed to get provider: ", err)
	}
	oidcProvider = provider

	// Configure OAuth2
	oauth2Config = oauth2.Config{
		ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		RedirectURL:  getEnv("REDIRECT_URL", "http://localhost:8080/callback"),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	// Setup Gin router
	r := gin.Default()

	// Load HTML template
	r.LoadHTMLGlob("templates/*")

	// Routes
	r.GET("/", indexHandler)
	r.GET("/login", loginHandler)
	r.GET("/callback", callbackHandler)
	r.GET("/profile", profileHandler)

	// Start server
	fmt.Println("Server starting on :8080")
	fmt.Println("Make sure to set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET environment variables")
	r.Run(":8080")
}

func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "OAuth POC",
	})
}

func loginHandler(c *gin.Context) {
	url := oauth2Config.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func callbackHandler(c *gin.Context) {
	// Verify state
	if c.Query("state") != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	// Exchange code for token
	code := c.Query("code")
	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	// Verify ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No id_token found"})
		return
	}

	verifier := oidcProvider.Verifier(&oidc.Config{ClientID: oauth2Config.ClientID})
	idToken, err := verifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify ID token"})
		return
	}

	// Extract claims
	var claims map[string]interface{}
	idToken.Claims(&claims)

	// Store user info in session (simplified - using query params for demo)
	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/profile?name=%s&email=%s",
		claims["name"], claims["email"]))
}

func profileHandler(c *gin.Context) {
	name := c.Query("name")
	email := c.Query("email")

	if name == "" || email == "" {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "OAuth POC",
			"error": "Please login first",
		})
		return
	}

	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Profile</title>
		<script src="https://cdn.tailwindcss.com"></script>
	</head>
	<body class="bg-gray-100 min-h-screen flex items-center justify-center">
		<div class="bg-white p-8 rounded-lg shadow-md max-w-md w-full">
			<h1 class="text-2xl font-bold text-green-600 mb-4">Login Successful!</h1>
			<div class="space-y-2">
				<p><strong>Name:</strong> ` + name + `</p>
				<p><strong>Email:</strong> ` + email + `</p>
			</div>
			<a href="/" class="mt-4 inline-block bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600">
				Back to Home
			</a>
		</div>
	</body>
	</html>`

	c.Data(http.StatusOK, "text/html", []byte(html))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
