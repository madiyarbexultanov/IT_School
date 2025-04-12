package models

import (
	"time"

	"github.com/google/uuid"
)

type LessonsFilters struct {
	PaymentStatus string
	LessonsStatus string
}

type Lessons struct {
	Id            uuid.UUID  `json:"id"`
	StudentId     uuid.UUID  `json:"student_id"`
	Date          *time.Time `json:"date"`
	Feedback      string     `json:"feedback"`
	PaymentStatus string     `json:"payment_status"`
	LessonsStatus string     `json:"lessons_status"`
	FeedbackDate  *time.Time `json:"feedback_date"`
	CreatedAt     *time.Time `json:"created_at"`
}
