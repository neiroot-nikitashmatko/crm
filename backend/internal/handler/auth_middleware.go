package handler

import (
	"net/http"
	"strings"

	"proclients/backend/internal/auth"
)

func isPublicPath(path string) bool {
	switch path {
	case "/health", "/api/v1/auth/login":
		return true
	default:
		return false
	}
}

func withAuth(jwtManager *auth.Manager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions || isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		header := strings.TrimSpace(r.Header.Get("Authorization"))
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "authorization required")
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		claims, err := jwtManager.Parse(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := auth.WithClaims(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
