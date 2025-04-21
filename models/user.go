package models

import "time"

type User struct {
	Id                  int
	Full_name			string
	Email               string
	PasswordHash        string
	Telephone			string
	RoleID              int
	ResetTokenExpiresAt time.Time
}