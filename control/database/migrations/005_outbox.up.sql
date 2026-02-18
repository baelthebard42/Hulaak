-- +migrate Up

CREATE TYPE delivery_status_nats AS ENUM(
  'null', 'processing', 'success'
);

CREATE TABLE outbox(

id UUID PRIMARY KEY,
delivery_id UUID NOT NULL UNIQUE,
created_at TIMESTAMP NOT NULL DEFAULT now(),
published_at TIMESTAMP NULL,
status delivery_status_nats NOT NULL DEFAULT 'null',

FOREIGN KEY (delivery_id) REFERENCES delivery(id)

);