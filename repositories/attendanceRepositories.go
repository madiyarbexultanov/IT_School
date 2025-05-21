package repositories

import (
	"context"
	"it_school/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AttendanceRepository struct {
	db *pgxpool.Pool
}

func NewAttendanceRepository(conn *pgxpool.Pool) *AttendanceRepository {
	return &AttendanceRepository{db: conn}
}

func (r *AttendanceRepository) CreateAttendance(c context.Context,attendance *models.Attendance,lesson *models.AttendanceLesson,
	freeze *models.AttendanceFreeze, prolongation *models.AttendanceProlongation,) (uuid.UUID, error) {

	tx, err := r.db.Begin(c)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(c)

	attendance.ID = uuid.New()

	_, err = tx.Exec(c, `
		INSERT INTO attendance (id, student_id, course_id, type)
		VALUES ($1, $2, $3, $4)
	`, attendance.ID, attendance.StudentId, attendance.CourseId, attendance.Type)
	if err != nil {
		return uuid.Nil, err
	}

	switch attendance.Type {
	case "урок":
		_, err = tx.Exec(c, `
			INSERT INTO attendance_lessons (attendance_id, curator_id, date, format, feedback, feedbackdate, lessons_status)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, attendance.ID, lesson.CuratorId, lesson.Date, lesson.Format, lesson.Feedback, lesson.FeedbackDate, lesson.LessonStatus)

	case "заморозка":
		_, err = tx.Exec(c, `
			INSERT INTO attendance_freezes (attendance_id, start_date, end_date, comment)
			VALUES ($1, $2, $3, $4)
		`, attendance.ID, freeze.StartDate, freeze.EndDate, freeze.Comment)

	case "пролонгация":
		_, err = tx.Exec(c, `
			INSERT INTO attendance_prolongations (attendance_id, payment_type, date, amount, comment)
			VALUES ($1, $2, $3, $4, $5)
		`, attendance.ID, prolongation.PaymentType, prolongation.Date, prolongation.Amount, prolongation.Comment)
	}

	if err != nil {
		return uuid.Nil, err
	}

	if err = tx.Commit(c); err != nil {
		return uuid.Nil, err
	}

	return attendance.ID, nil
}

func (r *AttendanceRepository) FindByStudent(c context.Context, studentID uuid.UUID) ([]models.Attendance, error) {
	rows, err := r.db.Query(c, `
		SELECT id, student_id, course_id, type, created_at
		FROM attendance
		WHERE student_id = $1
		ORDER BY created_at DESC
	`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attendances []models.Attendance
	for rows.Next() {
		var a models.Attendance
		if err := rows.Scan(&a.ID, &a.StudentId, &a.CourseId, &a.Type, &a.CreatedAt); err != nil {
			return nil, err
		}
		attendances = append(attendances, a)
	}

	return attendances, nil
}

func (r *AttendanceRepository) Update(c context.Context, attendance *models.Attendance, lesson *models.AttendanceLesson, freeze *models.AttendanceFreeze, prolongation *models.AttendanceProlongation) error {
	tx, err := r.db.Begin(c)
	if err != nil {
		return err
	}
	defer tx.Rollback(c)

	_, err = tx.Exec(c, `
		UPDATE attendance SET student_id = $1, course_id = $2, type = $3 WHERE id = $4
	`, attendance.StudentId, attendance.CourseId, attendance.Type, attendance.ID)
	if err != nil {
		return err
	}

	switch attendance.Type {
	case "урок":
		_, err = tx.Exec(c, `
			UPDATE attendance_lessons
			SET curator_id = $1, date = $2, format = $3, feedback = $4, feedbackdate = $5, lessons_status = $6
			WHERE attendance_id = $7
		`, lesson.CuratorId, lesson.Date, lesson.Format, lesson.Feedback, lesson.FeedbackDate, lesson.LessonStatus, attendance.ID)

	case "заморозка":
		_, err = tx.Exec(c, `
			UPDATE attendance_freezes
			SET start_date = $1, end_date = $2, comment = $3
			WHERE attendance_id = $4
		`, freeze.StartDate, freeze.EndDate, freeze.Comment, attendance.ID)

	case "пролонгация":
		_, err = tx.Exec(c, `
			UPDATE attendance_prolongations
			SET payment_type = $1, date = $2, amount = $3, comment = $4
			WHERE attendance_id = $5
		`, prolongation.PaymentType, prolongation.Date, prolongation.Amount, prolongation.Comment, attendance.ID)
	}

	if err != nil {
		return err
	}

	return tx.Commit(c)
}

func (r *AttendanceRepository) Delete(c context.Context, attendanceID uuid.UUID) error {
	_, err := r.db.Exec(c, `DELETE FROM attendance WHERE id = $1`, attendanceID)
	return err
}

func (r *AttendanceRepository) Exists(c context.Context, id uuid.UUID) (bool, error) {
    var exists bool
    err := r.db.QueryRow(c, "SELECT EXISTS(SELECT 1 FROM attendance WHERE id = $1)", id).Scan(&exists)
    return exists, err
}