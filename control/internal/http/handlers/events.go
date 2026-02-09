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

	source_username := ctx.Value("username")

	e, err := h.service.PostEvent(ctx, req.Event_Type, source_username.(string), req.Event_Destination, req.Payload)

	if err != nil {
		log.Printf("Failed to enter to database: %v", err)
		http.Error(w, "Error entering event to database, please try again", http.StatusBadRequest)
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(e)
}
