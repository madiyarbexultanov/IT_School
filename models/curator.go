package models

import "github.com/google/uuid"

type Curator struct {
	UserID     uuid.UUID   `json:"user_id"`
	StudentIDs []uuid.UUID `json:"student_ids"`
	CourseIDs  []uuid.UUID `json:"course_ids"`
}
