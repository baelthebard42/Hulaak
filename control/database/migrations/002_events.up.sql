-- +migrate Up

CREATE TABLE events(
    event_id UUID PRIMARY KEY,
    event_type VARCHAR(255),

    event_source UUID NOT NULL,
    FOREIGN KEY (event_source) REFERENCES client_user(client_id),
    payload JSONB,
    received_at TIMESTAMP DEFAULT now()
);