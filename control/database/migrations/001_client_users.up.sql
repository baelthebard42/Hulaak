-- +migrate Up

CREATE TABLE client_user(
    client_id UUID PRIMARY KEY,
    client_username VARCHAR(255) UNIQUE,
    client_email VARCHAR(254) UNIQUE,
    password_hash TEXT not null,
    created_at TIMESTAMP DEFAULT now()

);

