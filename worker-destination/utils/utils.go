package utils

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/avast/retry-go"
	_ "github.com/lib/pq"
)

type Payload struct {
	Delivery_id string `json:"delivery_id"`
}

type Webhook struct {
	EventID     string          `json:"event_id"`
	EventType   string          `json:"event_type"`
	EventSource string          `json:"event_source"`
	Payload     json.RawMessage `json:"payload"`
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

func FetchDeliveryDetails(db *sql.DB, delivery_id string) (*Webhook, string, error) {

	var endpoint string
	var webhook Webhook

	err := db.QueryRow(`

    SELECT 
    events.event_id, 
    events.event_type, 
    events.event_source, 
    events.payload, 
    endpoints.endpoint_url
	FROM delivery
	INNER JOIN events 
		ON delivery.event_id = events.event_id
	INNER JOIN endpoints 
		ON delivery.endpoint_id = endpoints.endpoint_id
	WHERE delivery.id = $1;

	`, delivery_id).Scan(&webhook.EventID, &webhook.EventType, &webhook.EventSource, &webhook.Payload, &endpoint)

	if err != nil {
		log.Fatalln("Unable to retrieve details for delivery %v: %v", delivery_id, err)
		return nil, "", err
	}

	return &webhook, endpoint, nil

}

func UpdateAfterRetry(db *sql.DB, delivery_id string) error {

	_, err := db.Exec(`
	UPDATE delivery
	SET 
		num_attempts = COALESCE(num_attempts, 0) + 1,
		last_attempt_at = now()
	WHERE id = $1;
	`, delivery_id)

	return err

}

func UpdateAfterSuccess(db *sql.DB, delivery_id string) error {

	_, err := db.Exec(`
	UPDATE delivery
	SET 
		num_attempts = COALESCE(num_attempts, 0) + 1,
		last_attempt_at = now()
		status = 'success'
	WHERE id = $1;
	`, delivery_id)

	return err

}
