package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/baelthebard42/Hulaak/control/internal/events"
)

type EventHandler struct {
	service events.EventService
}

func NewEventHandler(service events.EventService) *EventHandler {
	return &EventHandler{service: service}
}

type ReceiveEventRequest struct {
	Event_Type        string          `json:"event_type"`
	Event_Destination string          `json:"event_destination"`
	Payload           json.RawMessage `json:"payload"`
}

type PostEndpointRequest struct {
	DestinationRef string `json:"destination_ref"`
	EventType      string `json:"event_type"`
	Endpoint       string `json:"endpoint"`
}

func (h *EventHandler) ReceiveEvent(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ReceiveEventRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Event_Destination == "" || req.Event_Type == "" || req.Payload == nil {

		http.Error(w, "Certain fields are missing. Must include: destination, type, payload", http.StatusBadRequest)
		return

	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := ctx.Value("userID")

	e, err := h.service.PostEvent(ctx, req.Event_Type, userID.(string), req.Event_Destination, req.Payload)

	if err != nil {
		log.Printf("Failed to enter to database: %v", err)
		http.Error(w, "Error entering event to database, please try again", http.StatusBadRequest)
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(e)
}

func (h *EventHandler) PostEndpoint(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PostEndpointRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.DestinationRef == "" || req.EventType == "" || req.Endpoint == "" {

		http.Error(w, "Certain fields are missing. Must include: destination_ref, event_type, endpoint where the event must be sent", http.StatusBadRequest)
		return

	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err := h.service.PostEndpoint(ctx, req.DestinationRef, req.EventType, req.Endpoint)

	if err != nil {
		log.Printf("Failed to enter database: %v", err)
		http.Error(w, "The endpoint could not be registered", http.StatusBadRequest)
		return

	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return

}
