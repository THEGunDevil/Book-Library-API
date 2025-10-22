package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	PhoneNumber     string `json:"phone_number"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Bio             string `json:"bio"`        // added
	Role            string `json:"role"`
	TokenVersion    int    `json:"token_version"` // added
}

type UpdateUserRequest struct {
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	Bio         *string `json:"bio"`          // added
	PhoneNumber *string `json:"phone_number"`
}

type UserResponse struct {
	ID           uuid.UUID `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Bio          string    `json:"bio"`          // added
	Email        string    `json:"email"`
	PhoneNumber  string    `json:"phone_number"`
	CreatedAt    time.Time `json:"created_at"`
	Role         string    `json:"role"`
}
