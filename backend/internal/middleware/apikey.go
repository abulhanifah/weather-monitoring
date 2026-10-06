package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"

	"github.com/abulhanifah/weather-monitoring/internal/constants"
	"github.com/abulhanifah/weather-monitoring/internal/repositories"
	"gorm.io/gorm"
)

type ctxKey string

const deviceIDKey ctxKey = "device_id"

// DeviceIDFromContext ambil device id yang diautentikasi via API key.
func DeviceIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(deviceIDKey).(string)
	return id, ok && id != ""
}

// APIKeyAuthMiddleware validasi X-API-Key milik device untuk route ingest.
// Key di-hash (sha256 hex) lalu dicocokkan ke api key aktif (belum revoked).
func APIKeyAuthMiddleware(repo *repositories.DeviceRepository, prefix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			rawKey := strings.TrimSpace(r.Header.Get("X-API-Key"))
			if rawKey == "" || !strings.HasPrefix(rawKey, prefix) {
				slog.WarnContext(ctx, "invalid api key", slog.String("error_code", constants.ErrCodeInvalidAPIKey))
				http.Error(w, `{"error": "Unauthorized: Invalid API Key"}`, http.StatusUnauthorized)
				return
			}

			hashBytes := sha256.Sum256([]byte(rawKey))
			meta, err := repo.FindAPIKeyByHash(ctx, hex.EncodeToString(hashBytes[:]))
			if err != nil {
				code := constants.ErrCodeInvalidAPIKey
				if err != gorm.ErrRecordNotFound {
					slog.ErrorContext(ctx, "api key lookup failed", slog.Any("error", err.Error()))
				}
				slog.WarnContext(ctx, "invalid api key", slog.String("error_code", code))
				http.Error(w, `{"error": "Unauthorized: Invalid API Key"}`, http.StatusUnauthorized)
				return
			}

			if meta.IsRevoked {
				slog.WarnContext(ctx, "revoked api key", slog.String("error_code", constants.ErrCodeRevokedAPIKey))
				http.Error(w, `{"error": "Unauthorized: Revoked API Key"}`, http.StatusUnauthorized)
				return
			}

			if _, err := repo.FindByID(ctx, meta.DeviceID); err != nil {
				slog.WarnContext(ctx, "api key device not found", slog.String("error_code", constants.ErrCodeInvalidAPIKey))
				http.Error(w, `{"error": "Unauthorized: Invalid API Key"}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, deviceIDKey, meta.DeviceID)))
		})
	}
}
