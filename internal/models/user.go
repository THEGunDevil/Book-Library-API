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
	Bio             string `json:"bio"` // added
	Role            string `json:"role"`
	TokenVersion    int    `json:"token_version"` // added
	IsBanned        bool   `json:"is_banned"`
	BanReason       string `json:"ban_reason"`
	BanUntil        string `json:"ban_until"`        // optional, RFC3339 format
	IsPermanentBan  bool   `json:"is_permanent_ban"` // true = permanent ban

}

type UpdateUserRequest struct {
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	Bio         *string `json:"bio"` // added
	PhoneNumber *string `json:"phone_number"`
}

type UserResponse struct {
	ID                  uuid.UUID  `json:"id"`
	FirstName           string     `json:"first_name"`
	LastName            string     `json:"last_name"`
	Bio                 string     `json:"bio"` // added
	Email               string     `json:"email"`
	PhoneNumber         string     `json:"phone_number"`
	CreatedAt           time.Time  `json:"created_at"`
	Role                string     `json:"role"`
	TokenVersion        int        `json:"token_version"` // added
	IsBanned            bool       `json:"is_banned"`
	BanReason           string     `json:"ban_reason"`
	BanUntil            *time.Time `json:"ban_until,omitempty"` // nil = no ban date
	IsPermanentBan      bool       `json:"is_permanent_ban"`
	AllBorrowsCount    int        `json:"all_borrows_count,omitempty"`    // true = permanent ban
	ActiveBorrowsCount int        `json:"active_borrows_count,omitempty"` // true = permanent ban
}
type BanRequest struct {
	IsBanned       bool       `json:"is_banned"`
	BanReason      string     `json:"ban_reason"`
	BanUntil       *time.Time `json:"ban_until"`        // nullable
	IsPermanentBan bool       `json:"is_permanent_ban"` // true = permanent ban
}
