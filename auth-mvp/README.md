# OAuth + OIDC POC

Simple Google OAuth 2.0 + OpenID Connect demo with Go backend and HTMX frontend.

## Setup

1. **Get Google OAuth credentials:**
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select existing
   - Enable Google+ API
   - Create OAuth 2.0 credentials
   - Set redirect URI: `http://localhost:8080/callback`

2. **Set environment variables:**
   ```bash
   export GOOGLE_CLIENT_ID="your_client_id"
   export GOOGLE_CLIENT_SECRET="your_client_secret"
   ```

3. **Run the server:**
   ```bash
   cd auth-mvp
   go mod tidy
   go run main.go
   ```

4. **Open browser:** http://localhost:8080

## Features

- ✅ Single file Go HTTP server (Gin framework)
- ✅ Tailwind CSS + HTMX frontend
- ✅ Google OAuth 2.0 + OIDC login
- ✅ Token verification with go-oidc
- ✅ User profile display

## Flow

1. `/` → Login page with Google button
2. `/login` → Redirect to Google OAuth
3. `/callback` → Handle OAuth callback
4. `/profile` → Show user info