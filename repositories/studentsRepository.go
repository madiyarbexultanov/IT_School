package repositories

import (
	"context"
	"it_school/logger"
	"it_school/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

	row := r.db.QueryRow(c, `INSERT INTO students(id, course_id, full_name, phone_number, parent_name, parent_phone_number, curator_id, courses, platform_link, crm_link, created_at) 
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
    RETURNING id`,
		student.Id,
		student.CourseId,
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

func (r *StudentsRepository) FindAll(c context.Context, filters models.StudentFilters) ([]models.Student, error) {
	sql := `SELECT 
	s.id,
	s.course_id, 
	s.full_name, 
	s.phone_number, 
	s.parent_name, 
	s.parent_phone_number, 
	s.curator_id, 
	s.courses, 
	s.platform_link, 
	s.crm_link, 
	s.created_at,
	s.is_active
	FROM students s where 1=1`

	params := pgx.NamedArgs{}

	if filters.Search != "" {
		sql += " AND (s.full_name ILIKE @search OR s.email ILIKE @search)"
		params["search"] = "%" + filters.Search + "%"
	}

	if filters.Course != "" {
		sql += " AND @course = ANY(s.courses)"
		params["course"] = filters.Course
	}

	if filters.IsActive != "" {
		sql += " AND s.is_active = @is_active"
		isActive := filters.IsActive == "true"
		params["is_active"] = isActive
	}

	if filters.CuratorId != "" {
		sql += " AND s.curator_id = @curator_id"
		params["curator_id"] = filters.CuratorId
	}

	rows, err := r.db.Query(c, sql, params)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	students := make([]models.Student, 0)

	for rows.Next() {
		var student models.Student
		err := rows.Scan(
			&student.Id,
			&student.CourseId,
			&student.FullName,
			&student.PhoneNumber,
			&student.ParentName,
			&student.ParentPhoneNumber,
			&student.CuratorId,
			&student.Courses,
			&student.PlatformLink,
			&student.CrmLink,
			&student.CreatedAt,
			&student.IsActive,
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
			s.course_id,
			s.full_name, 
			s.phone_number, 
			s.parent_name, 
			s.parent_phone_number, 
       		s.curator_id, 
			s.courses, 
			s.platform_link, 
			s.crm_link, 
			s.created_at,
			s.is_active 
			FROM students s
			WHERE s.id = $1`

	var student models.Student
	l := logger.GetLogger()
	row := r.db.QueryRow(c, sql, studentId)
	err := row.Scan(
		&studentId,
		&student.CourseId,
		&student.FullName,
		&student.PhoneNumber,
		&student.ParentName,
		&student.ParentPhoneNumber,
		&student.CuratorId,
		&student.Courses,
		&student.PlatformLink,
		&student.CrmLink,
		&student.CreatedAt,
		&student.IsActive,
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
		course_id = $1,
		full_name = $2,
		phone_number = $3,
		parent_name = $4,
		parent_phone_number = $5,
		curator_id = $6,
		courses = $7,
		platform_link = $8,
		crm_link = $9,
		created_at = $10,
		is_active = $11
	WHERE id = $12`,
		updateStudents.CourseId,
		updateStudents.FullName,
		updateStudents.PhoneNumber,
		updateStudents.ParentName,
		updateStudents.ParentPhoneNumber,
		updateStudents.CuratorId,
		updateStudents.Courses,
		updateStudents.PlatformLink,
		updateStudents.CrmLink,
		updateStudents.CreatedAt,
		updateStudents.IsActive,
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
