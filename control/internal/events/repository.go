package events

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) PostEvent(ctx context.Context, e Event) (*Event, error) {

	err := r.db.QueryRowContext(ctx,
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

	return &e, nil
}
