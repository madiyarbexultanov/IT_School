package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
    Id                  uuid.UUID `json:"id"`
    Full_name           string    `json:"full_name"`
    Email               string    `json:"email"`
    PasswordHash        string    `json:"password_hash"`
    Telephone           string    `json:"telephone"`
    RoleID              uuid.UUID `json:"role_id"`
    ResetTokenExpiresAt time.Time `json:"reset_token_expires_at"`
}