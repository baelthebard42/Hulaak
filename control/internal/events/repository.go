package events

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	control_nats "github.com/baelthebard42/Hulaak/control/internal/nats"
	"github.com/google/uuid"
)

type DeliveryState struct {
	Delivery_id     string
	Num_attempts    int
	Last_attempt_at time.Time
	Status          string
	Last_error      string
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) PostEvent(ctx context.Context, e Event) (*Event, error) {

	// err := r.db.QueryRowContext(ctx,
	// 	`INSERT INTO events
	// 	(event_id, event_type, event_source, event_destination, payload)
	// 	 VALUES ($1, $2, $3, $4, $5)
	//  RETURNING received_at
	// 	`,
	// 	e.Event_ID,
	// 	e.Event_Type,
	// 	e.Event_Source,
	// 	e.Event_Destination,
	// 	e.Payload,
	// ).Scan(
	// 	&e.Received_At,
	// )

	//first check if a corresponding endpoint exists on the database for the event_destination, event_type

	var endpointId, endpoint_url string

	err := r.db.QueryRowContext(ctx, `
	SELECT endpoint_id, endpoint_url FROM endpoints
	WHERE destination_reference=$1 AND event_type=$2
	`, e.Event_Destination, e.Event_Type).Scan(&endpointId, &endpoint_url)

	if err != nil {

		return nil, err
	}

	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {

		return nil, err
	}

	defer tx.Rollback()

	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO events
		(event_id, event_type, event_source, event_destination, payload)
		 VALUES ($1, $2, $3, $4, $5)
     RETURNING received_at
		`,
		e.Event_ID,
		e.Event_Type,
		e.Event_Source,
		e.Event_Destination,
		e.Payload,
	).Scan(
		&e.Received_At,
	)

	if err != nil {

		return nil, err
	}

	delivery_id := uuid.New().String()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO delivery
		(id, event_id, endpoint_id)
		VALUES ($1, $2, $3)
		`,
		delivery_id, e.Event_ID, endpointId)

	if err != nil {

		return nil, err
	}

	_, err = tx.ExecContext(ctx, `
	INSERT INTO outbox
	(id, delivery_id, endpoint, event_source, event_type, payload)
	VALUES ($1, $2, $3, $4, $5, $6)
	`, uuid.New().String(), delivery_id, endpoint_url, e.Event_Source, e.Event_Type, e.Payload)

	if err != nil {

		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("tx commit: %w", err)
	}

	return &e, nil

}

func (r *Repository) PostEndpoint(ctx context.Context, destination_ref string, event_type string, endpoint string) error {

	userID := ctx.Value("userID")

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO endpoints
		(endpoint_id, destination_reference, source, event_type, endpoint_url)
		VALUES ($1, $2, $3, $4, $5)
		`,
		uuid.New().String(), destination_ref, userID, event_type, endpoint,
	)

	return err
}

func (r *Repository) UpdateAfterEvent(ctx context.Context, event control_nats.DeliveryResultEvent) error {
	query := `
		UPDATE delivery
		SET 
			status = COALESCE($1, status),
			last_error = $2,
			last_attempt_at = $3,
			num_attempts = num_attempts + 1
		WHERE id = $4
	`

	var (
		status    *string
		lastError string
	)

	if event.Succeeded {
		s := "success"
		status = &s
		lastError = ""
	} else {
		status = nil
		lastError = event.Error_message
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		status,
		lastError,
		event.Last_attempt,
		event.Delivery_id,
	)
	if err != nil {
		log.Println("error updating delivery table:", err)
		return err
	}

	return nil
}
func (r *Repository) GetDeliveryStateFromDID(ctx context.Context, delivery_id string) (DeliveryState, error) {

	var delivery_state DeliveryState

	err := r.db.QueryRowContext(ctx, `
	SELECT status, num_attempts, last_attempt_at, last_error FROM delivery
	WHERE id=$1
	`, delivery_id).Scan(&delivery_state.Status, &delivery_state.Num_attempts, &delivery_state.Last_attempt_at, &delivery_state.Last_error)

	if err != nil {
		log.Println("Error fetching delivery details: ", err)
		return DeliveryState{}, err
	}

	return delivery_state, nil

}
