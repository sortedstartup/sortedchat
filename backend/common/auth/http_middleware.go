package auth

import (
	"log/slog"
	"net/http"
	"strings"
)

// HTTPAuthMiddleware provides JWT authentication for HTTP handlers
type HTTPAuthMiddleware struct {
	validator    *JWTValidator
	skipPaths    map[string]bool
	skipPrefixes []string
	requireAuth  bool
}

// NewHTTPAuthMiddleware creates a new HTTP auth middleware
func NewHTTPAuthMiddleware(validator *JWTValidator, requireAuth bool) *HTTPAuthMiddleware {
	slog.Info("common:http_middleware:NewHTTPAuthMiddleware")
	return &HTTPAuthMiddleware{
		validator:   validator,
		skipPaths:   make(map[string]bool),
		requireAuth: requireAuth,
	}
}

// SkipPath adds a path to skip authentication (e.g., "/health", "/login")
func (m *HTTPAuthMiddleware) SkipPath(path string) {
	slog.Info("common:http_middleware:SkipPath", "path", path)
	m.skipPaths[path] = true
}

// SkipPaths adds multiple paths to skip authentication
func (m *HTTPAuthMiddleware) SkipPaths(paths []string) {
	slog.Info("common:http_middleware:SkipPaths", "paths", paths)
	for _, path := range paths {
		m.skipPaths[path] = true
	}
}

// SkipPrefix adds a path prefix to skip authentication (e.g., "/public/", "/auth/")
func (m *HTTPAuthMiddleware) SkipPrefix(prefix string) {
	slog.Info("common:http_middleware:SkipPrefix", "prefix", prefix)
	m.skipPrefixes = append(m.skipPrefixes, prefix)
}

// SkipPrefixes adds multiple path prefixes to skip authentication
func (m *HTTPAuthMiddleware) SkipPrefixes(prefixes []string) {
	slog.Info("common:http_middleware:SkipPrefixes", "prefixes", prefixes)
	m.skipPrefixes = append(m.skipPrefixes, prefixes...)
}

// Middleware returns an HTTP middleware function for JWT authentication
func (m *HTTPAuthMiddleware) Middleware(next http.Handler) http.Handler {
	slog.Debug("common:http_middleware:Middleware")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("common:http_middleware:Middleware", "path", r.URL.Path)
		// Check if this path should skip authentication
		if m.shouldSkipAuth(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract JWT token from request
		token, err := m.extractTokenFromRequest(r)
		if err != nil {
			slog.Error("common:http_middleware:Middleware", "path", r.URL.Path, "error", err)
			if m.requireAuth {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
			// If auth is not required, continue without user context
			next.ServeHTTP(w, r)
			return
		}

		// Validate token and extract user claims
		claims, err := m.validator.ValidateToken(token)
		if err != nil {
			slog.Error("common:http_middleware:Middleware", "path", r.URL.Path, "error", err)
			if m.requireAuth {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			// If auth is not required, continue without user context
			next.ServeHTTP(w, r)
			return
		}
		slog.Debug("common:http_middleware:Middleware", "path", r.URL.Path, "claims", claims)
		// Add user information to request context
		ctx := AddUserToContext(r.Context(), claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// MiddlewareFunc returns an HTTP middleware function (alternative signature)
func (m *HTTPAuthMiddleware) MiddlewareFunc(next http.HandlerFunc) http.HandlerFunc {
	slog.Debug("common:http_middleware:MiddlewareFunc")
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("common:http_middleware:MiddlewareFunc", "path", r.URL.Path)
		m.Middleware(next).ServeHTTP(w, r)
	}
}

// shouldSkipAuth checks if authentication should be skipped for the given path
func (m *HTTPAuthMiddleware) shouldSkipAuth(path string) bool {
	slog.Debug("common:http_middleware:shouldSkipAuth", "path", path)
	// Check exact path matches
	if m.skipPaths[path] {
		return true
	}

	// Check prefix matches
	for _, prefix := range m.skipPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// extractTokenFromRequest extracts JWT token from HTTP request
func (m *HTTPAuthMiddleware) extractTokenFromRequest(r *http.Request) (string, error) {
	slog.Debug("common:http_middleware:extractTokenFromRequest")
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer "), nil
		}
		slog.Debug("common:http_middleware:extractTokenFromRequest", "authHeader", authHeader)
		return authHeader, nil
	}

	// Try custom headers
	if token := r.Header.Get("X-Auth-Token"); token != "" {
		return token, nil
	}

	if token := r.Header.Get("X-JWT-Token"); token != "" {
		return token, nil
	}

	// Try query parameter (less secure, but sometimes needed)
	if token := r.URL.Query().Get("token"); token != "" {
		return token, nil
	}

	// Try cookie (if you want to support cookie-based auth)
	if cookie, err := r.Cookie("jwt"); err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	if cookie, err := r.Cookie("auth_token"); err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	slog.Error("common:http_middleware:extractTokenFromRequest", "error", http.ErrNoCookie)

	return "", http.ErrNoCookie
}

// RequireRoleMiddleware creates HTTP middleware that requires specific roles
func RequireRoleMiddleware(roles ...string) func(http.Handler) http.Handler {
	slog.Debug("common:http_middleware:RequireRoleMiddleware", "roles", roles)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Debug("common:http_middleware:RequireRoleMiddleware", "path", r.URL.Path)
			userRoles, ok := GetUserRolesFromContext(r.Context())
			if !ok {
				slog.Error("common:http_middleware:RequireRoleMiddleware", "path", r.URL.Path, "error", "user not authenticated")
				http.Error(w, "User not authenticated", http.StatusUnauthorized)
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, requiredRole := range roles {
				for _, userRole := range userRoles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				slog.Error("common:http_middleware:RequireRoleMiddleware", "path", r.URL.Path, "error", "insufficient permissions")
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRoleFunc creates HTTP middleware function that requires specific roles
func RequireRoleFunc(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	slog.Debug("common:http_middleware:RequireRoleFunc", "roles", roles)
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			slog.Debug("common:http_middleware:RequireRoleFunc", "path", r.URL.Path)
			userRoles, ok := GetUserRolesFromContext(r.Context())
			if !ok {
				slog.Error("common:http_middleware:RequireRoleFunc", "path", r.URL.Path, "error", "user not authenticated")
				http.Error(w, "User not authenticated", http.StatusUnauthorized)
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, requiredRole := range roles {
				for _, userRole := range userRoles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				slog.Error("common:http_middleware:RequireRoleFunc", "path", r.URL.Path, "error", "insufficient permissions")
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next(w, r)
		}
	}
}
