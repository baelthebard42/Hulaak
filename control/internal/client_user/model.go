package client_user

import "time"

type User struct {
	Client_id       string
	Client_username string
	Email           string
	Password_hash   string
}

type UserResponse struct {
	Client_id       string    `json:"id"`
	Client_username string    `json:"username"`
	Email           string    `json:"email"`
	CreatedAt       time.Time `json:"created_at"`
}
