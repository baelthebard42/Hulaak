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

func ClaimOutboxBatch(db *sql.DB) ([]Payload, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
		UPDATE outbox
		SET status = 'processing'
		   
		WHERE id IN (
			SELECT id
			FROM outbox
			WHERE published_at IS NULL
			  
			ORDER BY created_at
			FOR UPDATE SKIP LOCKED
			LIMIT 50
		)
		RETURNING delivery_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []Payload

	for rows.Next() {
		var deliveryID string
		if err := rows.Scan(&deliveryID); err != nil {
			return nil, err
		}
		deliveries = append(deliveries, Payload{
			Delivery_id: deliveryID,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return deliveries, nil
}

func MarkAsPublished(db *sql.DB, deliveryID string) error {
	_, err := db.Exec(`
		UPDATE outbox
		SET published_at = NOW(),
		    status = 'success'
		WHERE delivery_id = $1
	`, deliveryID)

	return err
}
