package repositories

import (
	"context"
	"fmt"
	"students/logger"
	"students/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

//ТЗ
// Описание
// Куратор видит только своих учеников. и свои курсы
//(В запросах GET /lessons и GET /students нужно проверять curator_id, сравнивая его с id текущего пользователя.
//Если curator_id != текущий пользователь, API должно возвращать 403 Forbidden)
// Менеджер видит всех учеников и оплату уроков(Для него в GET /lessons добавляется фильтрация по статусу оплаты
// (оплачен / не оплачен / предоплата)
// Директор имеет полный доступ ко всему

//

type LessonsRepository struct {
	db *pgxpool.Pool
}

func NewLessonsRepository(conn *pgxpool.Pool) *LessonsRepository {
	return &LessonsRepository{db: conn}
}

func (r *LessonsRepository) Create(c context.Context, lessons models.Lessons) (uuid.UUID, error) {
	l := logger.GetLogger()
	lessons.Id = uuid.New()

	row := r.db.QueryRow(c, `INSERT INTO lessons (id, student_id, date, feedback, status, feedback_date, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7) 
RETURNING id`,
		lessons.Id,
		lessons.StudentId,
		lessons.Date,
		lessons.Feedback,
		lessons.Status,
		lessons.FeedbackDate,
		lessons.CreatedAt)

	err := row.Scan(&lessons.Id)
	if err != nil {
		l.Error("Ошибка при создании урока/предмета/дисциплины", zap.Error(err))
		return uuid.UUID{}, err
	}
	l.Info("Урок/предмет/дисциплина успешно создан", zap.String("id", lessons.Id.String())) ///тут нет название урока
	return lessons.Id, nil
}

func (r *LessonsRepository) FindById(c context.Context, lessonsId uuid.UUID) (models.Lessons, error) {
	sql := `select 
	l.id,
	l.student_id,
	l.date, 
	l.feedback,
	l.status,
	l.feedback_date,
	l.created_at
	from lessons l
	where l.id = $1
	`

	var lessons models.Lessons
	row := r.db.QueryRow(c, sql, lessonsId)
	if err := row.Scan(&lessons.Id,
		&lessons.StudentId,
		&lessons.Date,
		&lessons.Feedback,
		&lessons.Status,
		&lessons.FeedbackDate,
		&lessons.CreatedAt); err != nil {
		return models.Lessons{}, err
	}

	return lessons, nil
}

func (r *LessonsRepository) FindAll(c context.Context, filters models.LessonsFilters) ([]models.Lessons, error) {
	sql := `select 
	l.id,
	l.student_id,
	l.date, 
	l.feedback,
	l.status,
	l.feedback_date,
	l.created_at
	from lessons l
	where 1=1
	`
	//http://localhost:8081/lessons?status=не оплачен, http://localhost:8081/lessons?status=предоплата
	params := pgx.NamedArgs{}

	if filters.Status != "" {
		sql = fmt.Sprintf("%s and l.status = @status", sql)
		params["status"] = filters.Status
	}

	row, err := r.db.Query(c, sql, params)
	if err != nil {
		return nil, err
	}
	lessons := make([]models.Lessons, 0)

	for row.Next() {
		var lesson models.Lessons
		err := row.Scan(
			&lesson.Id,
			&lesson.StudentId,
			&lesson.Date,
			&lesson.Feedback,
			&lesson.Status,
			&lesson.FeedbackDate,
			&lesson.CreatedAt)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, lesson)
	}
	return lessons, nil
}

func (r *LessonsRepository) Update(c context.Context, updateLessons models.Lessons) error {
	l := logger.GetLogger()

	tx, err := r.db.Begin(c)
	if err != nil {
		l.Error("Ошибка начала транзакции", zap.String("db_msg", err.Error()))
		return err
	}

	defer func() {
		if p := recover(); p != nil || err != nil {
			l.Error("Откат транзакции", zap.Any("panic", p), zap.String("rollback_msg", err.Error()))
			tx.Rollback(c)
		}
	}()

	_, err = tx.Exec(c, `
	UPDATE lessons
	SET 
		student_id = $1,
		date = $2,
		feedback = $3,
		status = $4,
		feedback_date = $5,
		created_at = $6
	WHERE id = $7`,
		updateLessons.StudentId,
		updateLessons.Date,
		updateLessons.Feedback,
		updateLessons.Status,
		updateLessons.FeedbackDate,
		updateLessons.CreatedAt,
		updateLessons.Id)

	if err != nil {
		l.Error("Ошибка при обновлении", zap.String("db_msg", err.Error()))
		return err
	}

	err = tx.Commit(c)
	if err != nil {
		l.Error("Ошибка при коммите транзакции", zap.String("commit_msg", err.Error()))
		return err
	}

	l.Info("Занятия успешно обновлены", zap.String("lessonss_id", updateLessons.Id.String()))
	return nil
}

// id UUID PRIMARY KEY,
// student_id UUID REFERENCES students(id) ON DELETE CASCADE,
// date DATE NOT NULL,
// feedback TEXT,
// status ENUM('запланирован', 'проведен', 'пропущен', 'отменен'),
// feedback_date TIMESTAMP NOT NULL,
// created_at TIMESTAMP DEFAULT NOW()

func (r *LessonsRepository) Delete(c context.Context, lessonsId uuid.UUID) error {
	l := logger.GetLogger()

	tx, err := r.db.Begin(c)
	if err != nil {
		l.Error("Ошибка начала транзакции", zap.String("db_msg", err.Error()))
		return err
	}

	_, err = tx.Exec(c, "DELETE FROM lessons WHERE id = $1", lessonsId)
	if err != nil {
		l.Error(err.Error())
		tx.Rollback(c)
		return err
	}

	err = tx.Commit(c)
	if err != nil {
		l.Error("Ошибка при коммите транзакции", zap.String("commit_msg", err.Error()))
		return err
	}
	return nil
}
