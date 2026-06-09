package events

import "encoding/json"

type WebhookReceiveEvent struct {
	Delivery_id  string          `json:"delivery_id"`
	Event_type   string          `json:"event_type"`
	Event_source string          `json:"event_source"`
	Endpoint     string          `json:"endpoint"`
	Payload      json.RawMessage `json:"payload"`
}
