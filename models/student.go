package models

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	Id                uuid.UUID  `json:"id"`
	FullName          string     `json:"full_name"`
	PhoneNumber       *string    `json:"phone_number"`
	ParentName        string     `json:"parent_name"`
	ParentPhoneNumber *string    `json:"parent_phone_number"`
	CuratorId         *uuid.UUID `json:"curator_id"`
	Courses           []string   `json:"courses"`
	PlatformLink      string     `json:"platform_link"`
	CrmLink           string     `json:"crm_link"`
	CreatedAt         *time.Time `json:"created_at"`
}
