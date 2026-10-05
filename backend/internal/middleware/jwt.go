package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/abulhanifah/weather-monitoring/internal/constants"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuthMiddleware(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				slog.WarnContext(ctx, "missing or invalid jwt format", slog.String("error_code", constants.ErrCodeUnauthorized))
				http.Error(w, `{"error": "Unauthorized: Missing or Invalid Token"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method")
				}
				return []byte(secretKey), nil
			})

			if err != nil || !token.Valid {
				slog.WarnContext(ctx, "jwt verification failed", slog.Any("cause", err), slog.String("error_code", constants.ErrCodeUnauthorized))
				http.Error(w, `{"error": "Unauthorized: Invalid JWT Token"}`, http.StatusUnauthorized)
				return
			}

			// Ambil claims email
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				ctx = context.WithValue(ctx, "user_email", claims["email"])
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
