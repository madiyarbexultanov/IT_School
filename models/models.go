package models

import (
	"time"

	"github.com/google/uuid"
)

type LessonsFilters struct {
	Status     string

}

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

type Lessons struct {
	Id           uuid.UUID  `json:"id"`
	StudentId    uuid.UUID  `json:"student_id"`
	Date         *time.Time `json:"date"`
	Feedback     string     `json:"feedback"`
	Status       string     `json:"status"`
	FeedbackDate *time.Time `json:"feedback_date"`
	CreatedAt    *time.Time `json:"created_at"`
}
