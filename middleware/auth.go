package middleware

import (
	"context"
	"diabetes-agent-mcp-server/config"
	"fmt"
	"log/slog"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Claims struct {
	UserEmail string `json:"email"`
	jwt.RegisteredClaims
}

func AuthMiddleware(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			return nil, fmt.Errorf("missing authorization header")
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return nil, fmt.Errorf("invalid authorization header format")
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := validateToken(token)
		if err != nil {
			return nil, fmt.Errorf("invalid token: %w", err)
		}

		// 将用户邮箱添加到上下文
		ctx = context.WithValue(ctx, "user_email", claims.UserEmail)

		return next(ctx, req)
	}
}

func validateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Cfg.JWT.SecretKey), nil
	})

	if err != nil || !token.Valid {
		slog.Info("Invalid token",
			"user_email", claims.UserEmail,
			"err", err,
		)
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
