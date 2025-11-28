package auth

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	slog.Debug("common:auth:jwt:NewJWTValidator")
	return &JWTValidator{
		secret:    secret,
		issuer:    issuer,
		algorithm: "HS256",
	}
}

// ValidateToken validates a JWT token and returns user claims
func (v *JWTValidator) ValidateToken(tokenString string) (*UserClaims, error) {
	slog.Debug("common:auth:jwt:ValidateToken")
	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	if tokenString == "" {
		slog.Error("common:auth:jwt:ValidateToken", "error", "empty token")
		return nil, fmt.Errorf("empty token")
	}

	// Parse and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		slog.Debug("common:auth:jwt:ValidateToken")
		// Validate the signing method
		if token.Method.Alg() != v.algorithm {
			slog.Error("common:auth:jwt:ValidateToken", "error", "unexpected signing method", "method", token.Method.Alg())
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.secret, nil
	})

	if err != nil {
		slog.Error("common:auth:jwt:ValidateToken", "message", "failed to parse token", "error", err)
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		slog.Error("common:auth:jwt:ValidateToken", "error", "invalid token")
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		slog.Error("common:auth:jwt:ValidateToken", "error", "invalid token claims")
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate issuer
	if iss, ok := claims["iss"].(string); !ok || iss != v.issuer {
		slog.Error("common:auth:jwt:ValidateToken", "error", "invalid issuer", "iss", iss, "expected", v.issuer)
		return nil, fmt.Errorf("invalid issuer")
	}

	// Extract user information
	userClaims := &UserClaims{}

	if sub, ok := claims["sub"].(string); ok {
		userClaims.UserID = sub
	} else {
		slog.Error("common:auth:jwt:ValidateToken", "error", "missing or invalid subject claim")
		return nil, fmt.Errorf("missing or invalid subject claim")
	}

	if email, ok := claims["email"].(string); ok {
		userClaims.Email = email
	} else {
		slog.Error("common:auth:jwt:ValidateToken", "error", "missing or invalid email claim")
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
	slog.Debug("common:jwt:AddUserToContext")
	ctx = context.WithValue(ctx, UserClaimsKey, claims)
	ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
	ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
	ctx = context.WithValue(ctx, UserRolesKey, claims.Roles)
	return ctx
}

// GetUserFromContext extracts user claims from context
func GetUserFromContext(ctx context.Context) (*UserClaims, bool) {
	slog.Debug("common:jwt:GetUserFromContext")
	claims, ok := ctx.Value(UserClaimsKey).(*UserClaims)
	return claims, ok
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	slog.Debug("common:jwt:GetUserIDFromContext")
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

func GetUserIDFromContext_WithError(ctx context.Context) (string, error) {
	slog.Debug("common:jwt:GetUserIDFromContext_WithError")
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		slog.Error("common:jwt:GetUserIDFromContext_WithError", "error", "user ID not found")
		return "", status.Errorf(codes.Unauthenticated, "user ID not found")
	}
	return userID, nil
}

// GetUserEmailFromContext extracts user email from context
func GetUserEmailFromContext(ctx context.Context) (string, bool) {
	slog.Debug("common:jwt:GetUserEmailFromContext")
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}

// GetUserRolesFromContext extracts user roles from context
func GetUserRolesFromContext(ctx context.Context) ([]string, bool) {
	slog.Debug("common:jwt:GetUserRolesFromContext")
	roles, ok := ctx.Value(UserRolesKey).([]string)
	return roles, ok
}

// HasRole checks if user has a specific role
func HasRole(ctx context.Context, role string) bool {
	slog.Debug("common:jwt:HasRole", "role", role)
	roles, ok := GetUserRolesFromContext(ctx)
	if !ok {
		slog.Error("common:jwt:HasRole", "error", "user roles not found")
		return false
	}

	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
