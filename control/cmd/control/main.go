package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/avast/retry-go"
	"github.com/baelthebard42/Hulaak/control/internal/client_user"
	"github.com/baelthebard42/Hulaak/control/internal/config"
	"github.com/baelthebard42/Hulaak/control/internal/db"
	"github.com/baelthebard42/Hulaak/control/internal/events"
	"github.com/baelthebard42/Hulaak/control/internal/http"
	"github.com/baelthebard42/Hulaak/control/internal/http/handlers"
	"github.com/baelthebard42/Hulaak/control/internal/http/routes"
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

	server := http.NewServer(router)

	if err := server.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}

}
