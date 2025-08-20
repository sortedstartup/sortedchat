package auth

import (
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
	return &HTTPAuthMiddleware{
		validator:   validator,
		skipPaths:   make(map[string]bool),
		requireAuth: requireAuth,
	}
}

// SkipPath adds a path to skip authentication (e.g., "/health", "/login")
func (m *HTTPAuthMiddleware) SkipPath(path string) {
	m.skipPaths[path] = true
}

// SkipPaths adds multiple paths to skip authentication
func (m *HTTPAuthMiddleware) SkipPaths(paths []string) {
	for _, path := range paths {
		m.skipPaths[path] = true
	}
}

// SkipPrefix adds a path prefix to skip authentication (e.g., "/public/", "/auth/")
func (m *HTTPAuthMiddleware) SkipPrefix(prefix string) {
	m.skipPrefixes = append(m.skipPrefixes, prefix)
}

// SkipPrefixes adds multiple path prefixes to skip authentication
func (m *HTTPAuthMiddleware) SkipPrefixes(prefixes []string) {
	m.skipPrefixes = append(m.skipPrefixes, prefixes...)
}

// Middleware returns an HTTP middleware function for JWT authentication
func (m *HTTPAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this path should skip authentication
		if m.shouldSkipAuth(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract JWT token from request
		token, err := m.extractTokenFromRequest(r)
		if err != nil {
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
			if m.requireAuth {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			// If auth is not required, continue without user context
			next.ServeHTTP(w, r)
			return
		}

		// Add user information to request context
		ctx := AddUserToContext(r.Context(), claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// MiddlewareFunc returns an HTTP middleware function (alternative signature)
func (m *HTTPAuthMiddleware) MiddlewareFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.Middleware(next).ServeHTTP(w, r)
	}
}

// shouldSkipAuth checks if authentication should be skipped for the given path
func (m *HTTPAuthMiddleware) shouldSkipAuth(path string) bool {
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
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer "), nil
		}
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

	return "", http.ErrNoCookie
}

// RequireRoleMiddleware creates HTTP middleware that requires specific roles
func RequireRoleMiddleware(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRoles, ok := GetUserRolesFromContext(r.Context())
			if !ok {
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
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRoleFunc creates HTTP middleware function that requires specific roles
func RequireRoleFunc(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			userRoles, ok := GetUserRolesFromContext(r.Context())
			if !ok {
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
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next(w, r)
		}
	}
}
