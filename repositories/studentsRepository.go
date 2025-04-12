package repositories

import (
	"context"
	"it_school/logger"
	"it_school/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type StudentsRepository struct {
	db *pgxpool.Pool
}

func NewStudentsRepository(conn *pgxpool.Pool) *StudentsRepository {
	return &StudentsRepository{db: conn}
}

func (r *StudentsRepository) Create(c context.Context, student models.Student) (uuid.UUID, error) {
	l := logger.GetLogger()
	student.Id = uuid.New()

	row := r.db.QueryRow(c, `INSERT INTO students(id, full_name, phone_number, parent_name, parent_phone_number, curator_id, courses, platform_link, crm_link, created_at) 
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
    RETURNING id`,
		student.Id,
		student.FullName,
		student.PhoneNumber,
		student.ParentName,
		student.ParentPhoneNumber,
		student.CuratorId,
		student.Courses,
		student.PlatformLink,
		student.CrmLink,
		student.CreatedAt,
	)

	err := row.Scan(&student.Id)
	if err != nil {
		l.Error("Ошибка при создании студента", zap.Error(err))
		return uuid.UUID{}, err
	}
	l.Info("Пользователь успешно создан", zap.String("id", student.Id.String()), zap.String("full_name", student.FullName))
	return student.Id, nil
}

func (r *StudentsRepository) FindAll(c context.Context) ([]models.Student, error) {
	sql := `SELECT 
	s.id, 
	s.full_name, 
	s.phone_number, 
	s.parent_name, 
	s.parent_phone_number, 
	s.curator_id, 
	s.courses, 
	s.platform_link, 
	s.crm_link, 
	s.created_at 
	FROM students s`

	rows, err := r.db.Query(c, sql)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	students := make([]models.Student, 0)

	for rows.Next() {
		var student models.Student
		err := rows.Scan(
			&student.Id,
			&student.FullName,
			&student.PhoneNumber,
			&student.ParentName,
			&student.ParentPhoneNumber,
			&student.CuratorId,
			&student.Courses,
			&student.PlatformLink,
			&student.CrmLink,
			&student.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		students = append(students, student)
	}
	return students, nil
}

func (r *StudentsRepository) FindById(c context.Context, studentId uuid.UUID) (models.Student, error) {
	sql := `SELECT 
			s.id, 
			s.full_name, 
			s.phone_number, 
			s.parent_name, 
			s.parent_phone_number, 
       		s.curator_id, 
			s.courses, 
			s.platform_link, 
			s.crm_link, 
			s.created_at 
			FROM students s
			WHERE s.id = $1`

	var student models.Student
	l := logger.GetLogger()
	row := r.db.QueryRow(c, sql, studentId)
	err := row.Scan(
		&studentId,
		&student.FullName,
		&student.PhoneNumber,
		&student.ParentName,
		&student.ParentPhoneNumber,
		&student.CuratorId,
		&student.Courses,
		&student.PlatformLink,
		&student.CrmLink,
		&student.CreatedAt,
	)
	if err != nil {
		l.Error("Ошибка запроса к базе", zap.String("db_msg", err.Error()))
		return models.Student{}, err
	}
	return student, nil
}

func (r *StudentsRepository) Update(c context.Context, updateStudents models.Student) error {
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
	UPDATE students
	SET 
		full_name = $1,
		phone_number = $2,
		parent_name = $3,
		parent_phone_number = $4,
		curator_id = $5,
		courses = $6,
		platform_link = $7,
		crm_link = $8,
		created_at = $9
	WHERE id = $10`,
		updateStudents.FullName,
		updateStudents.PhoneNumber,
		updateStudents.ParentName,
		updateStudents.ParentPhoneNumber,
		updateStudents.CuratorId,
		updateStudents.Courses,
		updateStudents.PlatformLink,
		updateStudents.CrmLink,
		updateStudents.CreatedAt,
		updateStudents.Id)

	if err != nil {
		l.Error("Ошибка при обновлении студента", zap.String("db_msg", err.Error()))
		return err
	}

	err = tx.Commit(c)
	if err != nil {
		l.Error("Ошибка при коммите транзакции", zap.String("commit_msg", err.Error()))
		return err
	}

	l.Info("Студент успешно обновлён", zap.String("student_id", updateStudents.Id.String()))
	return nil
}

func (r *StudentsRepository) Delete(c context.Context, studentId uuid.UUID) error {
	l := logger.GetLogger()

	tx, err := r.db.Begin(c)
	if err != nil {
		l.Error("Ошибка начала транзакции", zap.String("db_msg", err.Error()))
		return err
	}

	_, err = tx.Exec(c, "DELETE FROM students WHERE id = $1", studentId)
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
