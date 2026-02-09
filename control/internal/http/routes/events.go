package routes

import (
	"net/http"

	"github.com/baelthebard42/Hulaak/control/internal/http/handlers"
	"github.com/baelthebard42/Hulaak/control/internal/http/middleware"
)

func RegisterEventRoutes(h *handlers.EventHandler) func(*http.ServeMux) {

	return func(mux *http.ServeMux) {
		mux.Handle("/events", middleware.RequireAuth(http.HandlerFunc(h.ReceiveEvent)))
		mux.Handle("/endpoint", middleware.RequireAuth(http.HandlerFunc(h.PostEndpoint)))
	}
}
