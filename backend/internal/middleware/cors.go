package middleware

import (
	"net/http"
	"strings"
)

// CORSMiddleware tangani preflight OPTIONS dan header CORS.
// allowedOrigins: "*" atau daftar origin dipisah koma.
func CORSMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	allowed := parseOrigins(allowedOrigins)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowOrigin := "*"
			if origin != "" && len(allowed) > 0 && !allowed["*"] {
				if allowed[origin] {
					allowOrigin = origin
					w.Header().Set("Vary", "Origin")
				} else {
					// Origin tidak diizinkan: teruskan tanpa header CORS.
					if r.Method == http.MethodOptions {
						w.WriteHeader(http.StatusForbidden)
						return
					}
					next.ServeHTTP(w, r)
					return
				}
			}

			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func parseOrigins(s string) map[string]bool {
	out := map[string]bool{}
	for _, o := range strings.Split(s, ",") {
		if o = strings.TrimSpace(o); o != "" {
			out[o] = true
		}
	}
	return out
}
