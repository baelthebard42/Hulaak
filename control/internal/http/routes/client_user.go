package routes

import (
	"net/http"

	"github.com/baelthebard42/Hulaak/control/internal/http/handlers"
	"github.com/baelthebard42/Hulaak/control/internal/http/middleware"
)

func RegisterClientUserRoutes(
	h *handlers.ClientUserHandler,
) func(*http.ServeMux) {

	return func(mux *http.ServeMux) {
		mux.HandleFunc("/account", h.CreateAccountHandler)
		mux.HandleFunc("/login", h.LoginUserHandler)
		mux.Handle("/events", middleware.RequireAuth(http.HandlerFunc(handlers.ReceiveEvent)))
	}
}
