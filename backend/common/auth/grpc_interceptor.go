package auth

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GRPCAuthInterceptor provides JWT authentication for gRPC services
type GRPCAuthInterceptor struct {
	validator   *JWTValidator
	skipMethods map[string]bool
	requireAuth bool
}

// NewGRPCAuthInterceptor creates a new gRPC auth interceptor
func NewGRPCAuthInterceptor(validator *JWTValidator, requireAuth bool) *GRPCAuthInterceptor {
	slog.Debug("common:grpc_interceptor:NewGRPCAuthInterceptor")
	return &GRPCAuthInterceptor{
		validator:   validator,
		skipMethods: make(map[string]bool),
		requireAuth: requireAuth,
	}
}

// SkipMethod adds a method to skip authentication (e.g., "/health/check", "/auth/login")
func (i *GRPCAuthInterceptor) SkipMethod(method string) {
	slog.Debug("common:grpc_interceptor:SkipMethod", "method", method)
	i.skipMethods[method] = true
}

// SkipMethods adds multiple methods to skip authentication
func (i *GRPCAuthInterceptor) SkipMethods(methods []string) {
	slog.Debug("common:grpc_interceptor:SkipMethods", "methods", methods)
	for _, method := range methods {
		i.skipMethods[method] = true
	}
}

// UnaryInterceptor returns a gRPC unary server interceptor for JWT authentication
func (i *GRPCAuthInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	slog.Debug("common:grpc_interceptor:UnaryInterceptor")
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		slog.Debug("common:grpc_interceptor:UnaryInterceptor", "method", info.FullMethod)
		// Check if this method should skip authentication
		if i.skipMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// Extract JWT token from metadata
		token, err := i.extractTokenFromMetadata(ctx)
		if err != nil {
			slog.Error("common:grpc_interceptor:UnaryInterceptor", "method", info.FullMethod, "error", err)
			if i.requireAuth {
				return nil, status.Errorf(codes.Unauthenticated, "authentication required: %v", err)
			}
			// If auth is not required, continue without user context
			return handler(ctx, req)
		}

		// Validate token and extract user claims
		claims, err := i.validator.ValidateToken(token)
		if err != nil {
			slog.Error("common:grpc_interceptor:UnaryInterceptor", "method", info.FullMethod, "error", err)
			if i.requireAuth {
				return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
			}
			// If auth is not required, continue without user context
			return handler(ctx, req)
		}
		slog.Debug("common:grpc_interceptor:UnaryInterceptor", "method", info.FullMethod, "claims", claims)

		// Add user information to context
		ctx = AddUserToContext(ctx, claims)

		return handler(ctx, req)
	}
}

// StreamInterceptor returns a gRPC stream server interceptor for JWT authentication
func (i *GRPCAuthInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	slog.Debug("common:grpc_interceptor:StreamInterceptor")
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		slog.Debug("common:grpc_interceptor:StreamInterceptor", "method", info.FullMethod)
		// Check if this method should skip authentication
		if i.skipMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		// Extract JWT token from metadata
		token, err := i.extractTokenFromMetadata(ss.Context())
		if err != nil {
			slog.Error("common:grpc_interceptor:StreamInterceptor", "method", info.FullMethod, "error", err)
			if i.requireAuth {
				return status.Errorf(codes.Unauthenticated, "authentication required: %v", err)
			}
			// If auth is not required, continue without user context
			return handler(srv, ss)
		}

		// Validate token and extract user claims
		claims, err := i.validator.ValidateToken(token)
		if err != nil {
			slog.Error("common:grpc_interceptor:StreamInterceptor", "method", info.FullMethod, "error", err)
			if i.requireAuth {
				return status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
			}
			// If auth is not required, continue without user context
			return handler(srv, ss)
		}

		// Create a new context with user information
		ctx := AddUserToContext(ss.Context(), claims)
		slog.Debug("common:grpc_interceptor:StreamInterceptor", "method", info.FullMethod, "claims", claims)

		// Create a wrapped stream with the new context
		wrappedStream := &contextServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		return handler(srv, wrappedStream)
	}
}

// extractTokenFromMetadata extracts JWT token from gRPC metadata
func (i *GRPCAuthInterceptor) extractTokenFromMetadata(ctx context.Context) (string, error) {
	slog.Debug("common:grpc_interceptor:extractTokenFromMetadata")
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		slog.Error("common:grpc_interceptor:extractTokenFromMetadata", "error", "no metadata found")
		return "", fmt.Errorf("no metadata found")
	}

	// Try to get token from Authorization header
	authHeaders := md.Get("authorization")
	if len(authHeaders) > 0 {
		authHeader := authHeaders[0]
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer "), nil
		}
		return authHeader, nil
	}

	// Try to get token from custom header (e.g., "x-auth-token")
	tokenHeaders := md.Get("x-auth-token")
	if len(tokenHeaders) > 0 {
		return tokenHeaders[0], nil
	}

	// Try to get token from "jwt" header
	jwtHeaders := md.Get("jwt")
	if len(jwtHeaders) > 0 {
		return jwtHeaders[0], nil
	}

	slog.Error("common:grpc_interceptor:extractTokenFromMetadata", "error", "no authorization token found")

	return "", fmt.Errorf("no authorization token found")
}

// contextServerStream wraps grpc.ServerStream with a custom context
type contextServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the custom context
func (s *contextServerStream) Context() context.Context {
	return s.ctx
}

// RequireRole creates a gRPC interceptor that requires specific roles
func RequireRole(roles ...string) grpc.UnaryServerInterceptor {
	slog.Debug("common:grpc_interceptor:RequireRole", "roles", roles)
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		slog.Debug("common:grpc_interceptor:RequireRole", "method", info.FullMethod)
		userRoles, ok := GetUserRolesFromContext(ctx)
		if !ok {
			slog.Error("common:grpc_interceptor:RequireRole", "method", info.FullMethod, "error", "user not authenticated")
			return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
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
			slog.Error("common:grpc_interceptor:RequireRole", "method", info.FullMethod, "error", "insufficient permissions")
			return nil, status.Errorf(codes.PermissionDenied, "insufficient permissions")
		}

		return handler(ctx, req)
	}
}
