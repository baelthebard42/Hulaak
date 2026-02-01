package http

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(handler http.Handler) *Server {

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + port,
			Handler:      handler,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
}

func (s *Server) Run() error {
	errCh := make(chan error, 1)

	go func() {
		log.Printf("Server starting on %s", s.httpServer.Addr)
		errCh <- s.httpServer.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(ctx)

	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
