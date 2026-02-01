package client_user

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

func (r *Repository) CreateUser(ctx context.Context, u User) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO client_user 
		 (client_id, client_username, client_email, password_hash)
		 VALUES ($1, $2, $3, $4)`,
		u.Client_id,
		u.Client_username,
		u.Email,
		u.Password_hash,
	)
	return err
}

func (r *Repository) GetByUsername(
	ctx context.Context,
	username string,
) (*User, error) {

	row := r.db.QueryRowContext(
		ctx,
		`SELECT client_id, password_hash
		 FROM client_user
		 WHERE client_username = $1`,
		username,
	)

	var u User
	if err := row.Scan(
		&u.Client_id,
		&u.Email,
		&u.Password_hash,
	); err != nil {
		return nil, err
	}

	return &u, nil
}
