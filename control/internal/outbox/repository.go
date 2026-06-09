package outbox

import (
	"context"
	"database/sql"
	"encoding/json"

	control_nats "github.com/baelthebard42/Hulaak/control/internal/nats"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ClaimOutboxBatch(ctx context.Context, batch_size int) ([]control_nats.WebhookDeliveryEvent, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
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
		RETURNING delivery_id, endpoint, event_source, event_type, payload
	`, batch_size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []control_nats.WebhookDeliveryEvent

	for rows.Next() {
		var deliveryID, endpoint, event_source, event_type string
		var payload json.RawMessage
		if err := rows.Scan(&deliveryID, &endpoint, &event_source, &event_type, &payload); err != nil {
			return nil, err
		}
		deliveries = append(deliveries, control_nats.WebhookDeliveryEvent{
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

func (r *Repository) DeleteThisDelivery(ctx context.Context, deliveryID string) error {

	_, err := r.db.ExecContext(ctx, `
	DELETE FROM outbox
	WHERE delivery_id = $1
	`, deliveryID)

	return err
}

func (r *Repository) MarkAsUnprocessed(ctx context.Context, deliveryID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE outbox
		SET status = 'unprocessed'
		WHERE delivery_id = $1
	`, deliveryID)

	return err
}
