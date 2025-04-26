package models

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	Id                uuid.UUID  `json:"id"`
	CourseId          uuid.UUID  `json:"course_id"`
	FullName          string     `json:"full_name"`
	PhoneNumber       *string    `json:"phone_number"`
	ParentName        string     `json:"parent_name"`
	ParentPhoneNumber *string    `json:"parent_phone_number"`
	CuratorId         *uuid.UUID `json:"curator_id"`
	Courses           []string   `json:"courses"`
	PlatformLink      string     `json:"platform_link"`
	CrmLink           string     `json:"crm_link"`
	CreatedAt         *time.Time `json:"created_at"`
	IsActive          *string    `json:"is_active"`
}

type StudentFilters struct {
	Search    string
	Course    string
	IsActive  string
	CuratorId string
}
