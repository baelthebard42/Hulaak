package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/avast/retry-go"
	"github.com/baelthebard42/Hulaak/control/internal/client_user"
	"github.com/baelthebard42/Hulaak/control/internal/config"
	"github.com/baelthebard42/Hulaak/control/internal/db"
	"github.com/baelthebard42/Hulaak/control/internal/events"
	"github.com/baelthebard42/Hulaak/control/internal/http"
	"github.com/baelthebard42/Hulaak/control/internal/http/handlers"
	"github.com/baelthebard42/Hulaak/control/internal/http/routes"
	control_nats "github.com/baelthebard42/Hulaak/control/internal/nats"
	"github.com/baelthebard42/Hulaak/control/internal/outbox"
	"github.com/baelthebard42/Hulaak/control/internal/worker"
)

func main() {

	cfg := config.Load()

	var err error
	var postgres *sql.DB

	err = retry.Do(
		func() error {
			postgres, err = db.NewPostgres(cfg.DatabaseURL)
			if err != nil {
				log.Printf("failed to connect to DB, retrying: %v", err)
			}
			return err
		},
		retry.Delay(2*time.Second),
		retry.Attempts(10),
	)
	if err != nil {
		log.Fatalf("could not connect to DB: %v", err)
	}

	defer postgres.Close()

	userRepository := client_user.NewRepository(postgres)
	userService := client_user.NewClientUserService(*userRepository)
	userHandler := handlers.NewClientUserHandler(*userService)

	eventRepository := events.NewRepository(postgres)
	eventService := events.NewEventService(*eventRepository)
	eventHandler := handlers.NewEventHandler(*eventService)

	router := http.NewRouter(
		routes.RegisterClientUserRoutes(userHandler),
		routes.RegisterEventRoutes(eventHandler),
	)

	outboxRepository := outbox.NewRepository(postgres)
	NATS, err := control_nats.NewNATSConnection(cfg.NATSConnectionString)

	if err != nil {
		log.Fatalf("Error with NATS: %v", err)
	}

	server := http.NewServer(router)

	ctx, cancel := context.WithCancel(context.Background()) //context.Background is parent
	defer cancel()

	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM) // if OS sends SIGINT or SIGTERM, push it to sigch

	go func() {
		if err := server.Run(); err != nil {
			log.Printf("server stopped: %v", err)
			cancel()
		}
	}()

	go func() {
		worker.NewRunner(outboxRepository, NATS).Run(ctx)
	}()

	// when kubernetes deletes the pod, it sends SIGTERM which is pushed to sigch

	go func() {
		<-sigCh
		log.Println("shutdown signal received")
		cancel()
	}()

	<-ctx.Done()
}
