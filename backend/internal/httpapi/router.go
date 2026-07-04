package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/chef-stol/backend/internal/service"
)

// Handlers groups the domain handlers wired into the router.
type Handlers struct {
	Auth  *AuthHandler
	Slots *SlotsHandler
}

func NewRouter(h Handlers, authSvc *service.AuthService, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(CORS())
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(AccessLog(logger))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.Auth.Register)
		r.Post("/login", h.Auth.Login)
		r.With(RequireAuth(authSvc)).Post("/logout", h.Auth.Logout)
	})

	r.Route("/slots", func(r chi.Router) {
		r.Use(RequireAuth(authSvc))
		r.Get("/", h.Slots.List)
		r.Get("/{slotId}", h.Slots.Get)
	})

	return r
}
