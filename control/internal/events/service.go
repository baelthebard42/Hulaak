package events

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type EventService struct {
	repository Repository
}

func NewEventService(r Repository) *EventService {
	return &EventService{repository: r}
}

func (s *EventService) PostEvent(ctx context.Context, event_type string, event_source string, event_destination string, event_payload json.RawMessage) (*Event, error) {

	e := &Event{

		Event_ID:          uuid.New().String(),
		Event_Type:        event_type,
		Event_Source:      event_source,
		Event_Destination: event_destination,
		Payload:           event_payload,
		Received_At:       "", // this will be added from the database insert
	}

	e, err := s.repository.PostEvent(ctx, *e)

	if err != nil {
		return nil, err
	}

	return e, nil

}

func (s *EventService) PostEndpoint(ctx context.Context, destination_ref string, event_type string, endpoint string) error {
	err := s.repository.PostEndpoint(ctx, destination_ref, event_type, endpoint)

	return err
}
