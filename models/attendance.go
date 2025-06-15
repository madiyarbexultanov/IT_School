package models

import (
	"time"

	"github.com/google/uuid"
)

type Attendance struct {
	ID        uuid.UUID `json:"id"`
	StudentId uuid.UUID `json:"student_id"`
	CourseId  uuid.UUID `json:"course_id"`
	Type      string    `json:"type"` // lesson, freeze, prolongation
	CreatedAt time.Time `json:"created_at"`
}

type AttendanceLesson struct {
	AttendanceID  uuid.UUID  `json:"attendance_id"`
	CuratorId     uuid.UUID  `json:"curator_id"`
	Date          time.Time  `json:"date"`
	Format        *string    `json:"format"`
	Feedback      *string    `json:"feedback"`	
	LessonStatus  string     `json:"lessons_status"`
	CreatedAt 	  time.Time `json:"created_at"`	
	FeedbackDate  *time.Time `json:"feedback_date"`
}

type AttendanceFreeze struct {
	AttendanceID uuid.UUID `json:"attendance_id"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	Comment      *string   `json:"comment"`
}

type AttendanceProlongation struct {
	AttendanceID uuid.UUID `json:"attendance_id"`
	PaymentType  string    `json:"payment_type"`
	Date         time.Time `json:"date"`
	Amount       float64   `json:"amount"`
	Comment      *string   `json:"comment"`
}

type AttendanceFullResponse struct {
    Attendance  *Attendance           		`json:"attendance"`
    Lesson      *AttendanceLesson     		`json:"lesson,omitempty"`
    Freeze      *AttendanceFreeze     		`json:"freeze,omitempty"`
    Prolongation *AttendanceProlongation 	`json:"prolongation,omitempty"`
}