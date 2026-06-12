package worker_nats

import (
	"encoding/json"
	"time"
)

type WebhookReceiveEvent struct {
	Delivery_id  string          `json:"delivery_id"`
	Event_type   string          `json:"event_type"`
	Event_source string          `json:"event_source"`
	Endpoint     string          `json:"endpoint"`
	Payload      json.RawMessage `json:"payload"`
}

type DeliveryResultEvent struct {
	Delivery_id   string    `json:"delivery_id"`
	Succeeded     bool      `json:"succeeded"`
	Last_attempt  time.Time `json:"last_attempt"`
	Error_message string    `json:"error_message"`
}
