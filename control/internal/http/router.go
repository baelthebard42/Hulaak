package http

import "net/http"

type RouteRegistrar func(mux *http.ServeMux)

func NewRouter(registrars ...RouteRegistrar) http.Handler {
	mux := http.NewServeMux()

	for _, register := range registrars {
		register(mux)
	}

	return mux
}
