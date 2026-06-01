-- +migrate Up

DROP TYPE delivery_status_nats;
CREATE TYPE delivery_status_nats AS ENUM(
  'unprocessed', 'processing'
);

CREATE TABLE outbox(

id UUID PRIMARY KEY,
delivery_id UUID NOT NULL UNIQUE,
created_at TIMESTAMP NOT NULL DEFAULT now(),
status delivery_status_nats NOT NULL DEFAULT 'unprocessed',
endpoint VARCHAR(255) NOT NULL,
event_source UUID NOT NULL,
event_type VARCHAR(255),
payload JSONB,


FOREIGN KEY (delivery_id) REFERENCES delivery(id)

);