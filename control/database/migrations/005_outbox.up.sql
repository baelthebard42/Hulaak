-- +migrate Up

CREATE TABLE outbox(

id UUID PRIMARY KEY,
delivery_id UUID NOT NULL UNIQUE,
created_at TIMESTAMP NOT NULL DEFAULT now(),
published_at TIMESTAMP NULL,

FOREIGN KEY (delivery_id) REFERENCES delivery(id)

);