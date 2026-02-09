package events

import "encoding/json"

type Event struct {
	Event_ID          string          `json:"event_id"`
	Event_Type        string          `json:"event_type"`
	Event_Source      string          `json:"event_source"`
	Event_Destination string          `json:"event_destination"`
	Payload           json.RawMessage `json:"payload"`
	Received_At       string          `json:"received_at"`
}
