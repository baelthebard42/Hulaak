package events

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

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

	var endpointId string

	err := r.db.QueryRowContext(ctx, `
	SELECT endpoint_id FROM endpoints
	WHERE destination_reference=$1, event_type=$2
	`, e.Event_Destination, e.Event_Type).Scan(&endpointId)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
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
	(id, delivery_id)
	VALUES ($1, $2)
	`, uuid.New().String(), delivery_id)

	if err != nil {
		return nil, err
	}

	return &e, nil

}

func (r *Repository) PostEndpoint(ctx context.Context, destination_ref string, event_type string, endpoint string) error {

	userID := ctx.Value("userID")

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO endpoints
		(destination_reference, source, event_type, endpoint_url)
		VALUES ($1, $2, $3, $4)
		`,
		destination_ref, userID, event_type, endpoint,
	)

	return err
}
