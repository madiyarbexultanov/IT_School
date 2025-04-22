package models

import "github.com/google/uuid"

type Course struct {
	Id    uuid.UUID `json:"id"`
	Title string    `json:"title"`
}
