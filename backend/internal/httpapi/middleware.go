package httpapi

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/chef-stol/backend/internal/domain"
	"github.com/chef-stol/backend/internal/service"
)

// RequireAuth extracts "Authorization: Bearer <token>" and rejects with a contractual 401
// if it's missing or invalid (NFR-7). The auth/slots endpoints wired up so far only need to
// know the caller is authenticated, not who they are — per-client scoping (needed once
// bookings land, NFR-8) is deliberately not threaded through the request context yet.
func RequireAuth(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if !strings.HasPrefix(header, prefix) {
				mapDomainError(w, domain.ErrUnauthorized)
				return
			}
			if _, err := auth.Authenticate(r.Context(), strings.TrimPrefix(header, prefix)); err != nil {
				mapDomainError(w, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AccessLog logs one structured line per request.
func AccessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", middleware.GetReqID(r.Context()),
			)
		})
	}
}
