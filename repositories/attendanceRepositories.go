package repositories

import (
	"context"
	"database/sql"
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


func (r *AttendanceRepository) FindFullByStudent(ctx context.Context, studentID uuid.UUID) ([]models.AttendanceFullResponse, error) {
    rows, err := r.db.Query(ctx, `
        SELECT 
            a.id, a.student_id, a.course_id, a.type, a.created_at,

            -- lesson
            l.curator_id, l.date, l.format, l.feedback, l.lessons_status, l.feedbackdate,

            -- freeze
            f.start_date, f.end_date, f.comment,

            -- prolongation
            p.payment_type, p.date, p.amount, p.comment

        FROM attendance a
        LEFT JOIN attendance_lessons l ON a.id = l.attendance_id AND a.type = 'урок'
        LEFT JOIN attendance_freezes f ON a.id = f.attendance_id AND a.type = 'заморозка'
        LEFT JOIN attendance_prolongations p ON a.id = p.attendance_id AND a.type = 'пролонгация'
        WHERE a.student_id = $1
        ORDER BY a.created_at DESC
    `, studentID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var responses []models.AttendanceFullResponse

    for rows.Next() {
        var att models.Attendance

        // lesson (nullable)
        var lessonDate, feedbackDate sql.NullTime
        var format, feedback, lessonStatus sql.NullString
        var curatorID uuid.NullUUID

        // freeze (nullable)
        var startDate, endDate sql.NullTime
        var freezeComment sql.NullString

        // prolongation (nullable)
        var paymentType, prolongComment sql.NullString
        var prolongDate sql.NullTime
        var amount sql.NullFloat64

        err := rows.Scan(
            &att.ID, &att.StudentId, &att.CourseId, &att.Type, &att.CreatedAt,
            &curatorID, &lessonDate, &format, &feedback, &lessonStatus, &feedbackDate,
            &startDate, &endDate, &freezeComment,
            &paymentType, &prolongDate, &amount, &prolongComment,
        )
        if err != nil {
            return nil, err
        }

        response := models.AttendanceFullResponse{
            Attendance: &att,
        }

        switch att.Type {
        case "урок":
            lesson := models.AttendanceLesson{
                AttendanceID: att.ID,
            }

            hasData := false

            if curatorID.Valid {
                lesson.CuratorId = curatorID.UUID
                hasData = true
            }
            if lessonDate.Valid {
                lesson.Date = lessonDate.Time
                hasData = true
            }
            if format.Valid {
                lesson.Format = &format.String
                hasData = true
            }
            if feedback.Valid {
                lesson.Feedback = &feedback.String
                hasData = true
            }
            if feedbackDate.Valid {
                lesson.FeedbackDate = &feedbackDate.Time
                hasData = true
            }
            if lessonStatus.Valid {
                lesson.LessonStatus = lessonStatus.String
                hasData = true
            }

            if hasData {
                response.Lesson = &lesson
            }

        case "заморозка":
            freeze := models.AttendanceFreeze{
                AttendanceID: att.ID,
            }

            hasData := false

            if startDate.Valid {
                freeze.StartDate = startDate.Time
                hasData = true
            }
            if endDate.Valid {
                freeze.EndDate = endDate.Time
                hasData = true
            }
            if freezeComment.Valid {
                freeze.Comment = &freezeComment.String
                hasData = true
            }

            if hasData {
                response.Freeze = &freeze
            }

        case "пролонгация":
            prolongation := models.AttendanceProlongation{
                AttendanceID: att.ID,
            }

            hasData := false

            if paymentType.Valid {
                prolongation.PaymentType = paymentType.String
                hasData = true
            }
            if prolongDate.Valid {
                prolongation.Date = prolongDate.Time
                hasData = true
            }
            if amount.Valid {
                prolongation.Amount = amount.Float64
                hasData = true
            }
            if prolongComment.Valid {
                prolongation.Comment = &prolongComment.String
                hasData = true
            }

            if hasData {
                response.Prolongation = &prolongation
            }
        }

        responses = append(responses, response)
    }

    return responses, nil
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