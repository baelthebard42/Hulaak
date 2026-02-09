package handlers

import (
	"encoding/json"
	"net/http"
)

type EventRequest struct {
	EventType string `json:"event_type"`
}

type EventResponse struct {
	EventType string `json:"event_type"`
}

func ReceiveEvent(w http.ResponseWriter, r *http.Request) {

	var req EventRequest
	var resp EventResponse

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	print("The event type of received event is ", req.EventType)
	resp.EventType = req.EventType

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
