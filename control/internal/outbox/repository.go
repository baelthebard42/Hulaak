package outbox

import (
	"database/sql"
	"encoding/json"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type WebhookDetails struct {
	Delivery_id  string          `json:"delivery_id"`
	Event_type   string          `json:"event_type"`
	Event_source string          `json:"event_source"`
	Endpoint     string          `json:"endpoint"`
	Payload      json.RawMessage `json:"payload"`
}

func (r *Repository) ClaimOutboxBatch(batch_size int) ([]WebhookDetails, error) {
	tx, err := r.db.Begin()
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
			WHERE status='unprocessed'
			  
			ORDER BY created_at
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		)
		RETURNING delivery_id, endpoint, event_source, event_type, WebhookDetails
	`, batch_size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []WebhookDetails

	for rows.Next() {
		var deliveryID, endpoint, event_source, event_type string
		var payload json.RawMessage
		if err := rows.Scan(&deliveryID, &endpoint, &event_source, &event_type, &payload); err != nil {
			return nil, err
		}
		deliveries = append(deliveries, WebhookDetails{
			Delivery_id:  deliveryID,
			Endpoint:     endpoint,
			Event_source: event_source,
			Event_type:   event_type,
			Payload:      payload,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return deliveries, nil
}

func (r *Repository) DeleteThisDelivery(deliveryID string) error {

	_, err := r.db.Exec(`
	DELETE FROM outbox
	WHERE delivery_id = $1
	`, deliveryID)

	return err
}
