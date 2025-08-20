package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// UserClaims represents the user information extracted from JWT
type UserClaims struct {
	UserID string   `json:"sub"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	Issuer string   `json:"iss"`
}

// JWTValidator handles JWT token validation
type JWTValidator struct {
	secret    []byte
	issuer    string
	algorithm string
}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator(secret []byte, issuer string) *JWTValidator {
	return &JWTValidator{
		secret:    secret,
		issuer:    issuer,
		algorithm: "HS256",
	}
}

// ValidateToken validates a JWT token and returns user claims
func (v *JWTValidator) ValidateToken(tokenString string) (*UserClaims, error) {
	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	if tokenString == "" {
		return nil, fmt.Errorf("empty token")
	}

	// Parse and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if token.Method.Alg() != v.algorithm {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate issuer
	if iss, ok := claims["iss"].(string); !ok || iss != v.issuer {
		return nil, fmt.Errorf("invalid issuer")
	}

	// Extract user information
	userClaims := &UserClaims{}

	if sub, ok := claims["sub"].(string); ok {
		userClaims.UserID = sub
	} else {
		return nil, fmt.Errorf("missing or invalid subject claim")
	}

	if email, ok := claims["email"].(string); ok {
		userClaims.Email = email
	}

	if iss, ok := claims["iss"].(string); ok {
		userClaims.Issuer = iss
	}

	// Handle roles (can be []interface{} or []string)
	if rolesInterface, ok := claims["roles"]; ok {
		switch roles := rolesInterface.(type) {
		case []interface{}:
			for _, role := range roles {
				if roleStr, ok := role.(string); ok {
					userClaims.Roles = append(userClaims.Roles, roleStr)
				}
			}
		case []string:
			userClaims.Roles = roles
		}
	}

	return userClaims, nil
}

// Context keys for storing user information
type contextKey string

const (
	UserClaimsKey contextKey = "user_claims"
	UserIDKey     contextKey = "user_id"
	UserEmailKey  contextKey = "user_email"
	UserRolesKey  contextKey = "user_roles"
)

// AddUserToContext adds user claims to the context
func AddUserToContext(ctx context.Context, claims *UserClaims) context.Context {
	ctx = context.WithValue(ctx, UserClaimsKey, claims)
	ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
	ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
	ctx = context.WithValue(ctx, UserRolesKey, claims.Roles)
	return ctx
}

// GetUserFromContext extracts user claims from context
func GetUserFromContext(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*UserClaims)
	return claims, ok
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetUserEmailFromContext extracts user email from context
func GetUserEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}

// GetUserRolesFromContext extracts user roles from context
func GetUserRolesFromContext(ctx context.Context) ([]string, bool) {
	roles, ok := ctx.Value(UserRolesKey).([]string)
	return roles, ok
}

// HasRole checks if user has a specific role
func HasRole(ctx context.Context, role string) bool {
	roles, ok := GetUserRolesFromContext(ctx)
	if !ok {
		return false
	}

	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
