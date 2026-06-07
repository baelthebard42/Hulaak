package routes

import (
	"net/http"

	"github.com/baelthebard42/Hulaak/control/internal/http/handlers"
)

func RegisterClientUserRoutes(
	h *handlers.ClientUserHandler,
) func(*http.ServeMux) {

	return func(mux *http.ServeMux) {
		mux.HandleFunc("/healthz", h.HealthzHandler)
		mux.HandleFunc("/account", h.CreateAccountHandler)
		mux.HandleFunc("/login", h.LoginUserHandler)

	}
}
