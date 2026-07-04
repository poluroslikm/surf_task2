package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/chef-stol/backend/internal/domain"
	"github.com/chef-stol/backend/internal/service"
)

type contextKey int

const clientContextKey contextKey = iota

// RequireAuth extracts "Authorization: Bearer <token>" and rejects with a contractual 401
// if it's missing or invalid (NFR-7), then attaches the authenticated domain.Client to the
// request context so downstream handlers can scope reads/writes to the caller. Bookings needs
// this for per-client scoping (NFR-8: listBookings/getBooking/cancelBooking/submitRating) —
// auth/slots didn't, which is why this was deferred until now.
func RequireAuth(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if !strings.HasPrefix(header, prefix) {
				mapDomainError(w, domain.ErrUnauthorized)
				return
			}
			client, err := auth.Authenticate(r.Context(), strings.TrimPrefix(header, prefix))
			if err != nil {
				mapDomainError(w, err)
				return
			}
			ctx := context.WithValue(r.Context(), clientContextKey, client)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClientFromContext retrieves the domain.Client attached by RequireAuth. Only meaningful for
// handlers mounted behind RequireAuth — ok is false otherwise.
func ClientFromContext(ctx context.Context) (domain.Client, bool) {
	c, ok := ctx.Value(clientContextKey).(domain.Client)
	return c, ok
}

// CORS is a permissive dev-only CORS middleware: there's no fixed deployment origin yet (see
// README "Переменные окружения"), so it reflects whatever Origin the browser sent instead of
// hardcoding one. Must run before RequireAuth so preflight OPTIONS (sent by the browser ahead
// of any request carrying "Authorization") never hits the 401 check.
func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if origin := r.Header.Get("Origin"); origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
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
