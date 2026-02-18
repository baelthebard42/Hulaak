package main

import (
	"database/sql"
)

type payload struct {
	delivery_id string
}

func ClaimOutboxBatch(db *sql.DB) ([]payload, error) {
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

	var deliveries []payload

	for rows.Next() {
		var deliveryID string
		if err := rows.Scan(&deliveryID); err != nil {
			return nil, err
		}
		deliveries = append(deliveries, payload{
			delivery_id: deliveryID,
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
