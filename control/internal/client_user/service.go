package client_user

import (
	"context"

	"github.com/baelthebard42/Hulaak/control/internal/utils"
	"github.com/google/uuid"
)

type ClientUserService struct {
	repository Repository
}

func NewClientUserService(r Repository) *ClientUserService {
	return &ClientUserService{repository: r}
}

func (s *ClientUserService) CreateAccount(ctx context.Context, username string, email string, password string) (*User, error) {

	password_hash, err := utils.HashPassword(password)

	if err != nil {
		return nil, err
	}

	u := &User{
		Client_id:       uuid.New().String(),
		Client_username: username,
		Email:           email,
		Password_hash:   password_hash,
	}

	err = s.repository.CreateUser(ctx, *u)

	if err != nil {
		return nil, err
	}

	return u, nil

}
