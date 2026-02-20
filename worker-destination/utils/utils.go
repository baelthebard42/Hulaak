package utils

import (
	"database/sql"
	"log"
	"time"

	"github.com/avast/retry-go"
	_ "github.com/lib/pq"
)

type Payload struct {
	Delivery_id string `json:"delivery_id"`
}

func NewPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func NewDBConnection(connection_string string) (*sql.DB, error) {

	var err error
	var postgres *sql.DB

	err = retry.Do(
		func() error {
			log.Println("Connecting to database...")
			postgres, err = NewPostgres(connection_string)
			if err != nil {
				log.Println("failed to connect to DB, retrying: %v", err)
			}
			return err
		},
		retry.Delay(2*time.Second),
		retry.Attempts(10),
	)
	if err != nil {
		log.Fatalln("could not connect to DB: %v", err)
		return nil, err
	}

	log.Println("Worker-nats connected to database!")

	return postgres, nil

}
