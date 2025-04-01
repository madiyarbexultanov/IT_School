package models

import "time"

type User struct {
	Id                  int
	Email               string
	PasswordHash        string
	RoleID              int
	ResetTokenExpiresAt time.Time
}