-- +migrate Up

CREATE TYPE delivery_status AS ENUM(
  'pending', 'success', 'failed'
);

CREATE TABLE delivery(

  id UUID UNIQUE PRIMARY KEY,
  event_id UUID,
  endpoint_id UUID,
  status delivery_status NOT NULL DEFAULT 'pending',
  num_attempts INT,
  last_error TEXT,
  last_attempt_at TIMESTAMP,
  next_retry_at TIMESTAMP,
  created_at TIMESTAMP default now(),

  FOREIGN KEY (event_id) REFERENCES events(event_id),
  FOREIGN KEY (endpoint_id) REFERENCES endpoints(endpoint_id)
);

