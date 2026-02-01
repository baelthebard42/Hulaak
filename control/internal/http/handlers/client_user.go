package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/baelthebard42/Hulaak/control/internal/client_user"
)

type ClientUserHandler struct {
	service client_user.ClientUserService
}

func NewClientUserHandler(service client_user.ClientUserService) *ClientUserHandler {
	return &ClientUserHandler{service: service}
}

type CreateAccountRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateAccountResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func (h *ClientUserHandler) CreateAccountHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Username == "" || req.Password == "" {
		http.Error(w, "Email, username and password all must be provided in the body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	user, err := h.service.CreateAccount(ctx, req.Username, req.Email, req.Password)

	if err != nil {

		log.Printf("Failed to create account !! %v", err)
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return

	}

	resp := CreateAccountResponse{
		ID:       user.Client_id,
		Email:    user.Email,
		Username: user.Client_username,
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)

}
