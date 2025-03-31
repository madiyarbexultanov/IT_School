package models


type User struct {
	Id           int
	Email        string
	PasswordHash string
	RoleID       int
}