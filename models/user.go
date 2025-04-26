package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id                  uuid.UUID
	Full_name			string
	Email               string
	PasswordHash        string
	Telephone			string
	RoleID              uuid.UUID
	ResetTokenExpiresAt time.Time
}