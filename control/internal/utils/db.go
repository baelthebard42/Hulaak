package utils

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func MarkAsPublished(db *sql.DB, deliveryID string) error {
	_, err := db.Exec(`
		UPDATE outbox
		SET published_at = NOW(),
		    status = 'success'
		WHERE delivery_id = $1
	`, deliveryID)

	return err
}

func MarkAsNull(db *sql.DB, deliveryID string) error {
	_, err := db.Exec(`
		UPDATE outbox
		SET status = 'null'
		WHERE delivery_id = $1
	`, deliveryID)

	return err
}
