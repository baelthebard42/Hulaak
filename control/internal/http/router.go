package http

import "net/http"

type RouteRegistrar func(mux *http.ServeMux) // each registrar function should take one argument of *http.ServeMux type and add the necessary route to it

func NewRouter(registrars ...RouteRegistrar) http.Handler {
	mux := http.NewServeMux()

	for _, register := range registrars {
		register(mux)
	}

	return mux
}
