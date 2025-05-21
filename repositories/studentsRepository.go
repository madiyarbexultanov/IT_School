package repositories

import (
	"context"
	"it_school/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StudentsRepository struct {
	db *pgxpool.Pool
}

func NewStudentsRepository(conn *pgxpool.Pool) *StudentsRepository {
	return &StudentsRepository{db: conn}
}

func (r *StudentsRepository) Create(c context.Context, student models.Student) (uuid.UUID, error) {
	student.Id = uuid.New()

	row := r.db.QueryRow(c, `INSERT INTO students(id, course_id, full_name, phone_number, parent_name, parent_phone_number, curator_id, platform_link, crm_link, created_at, is_active) 
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
    RETURNING id`,
		student.Id,
		student.CourseId,
		student.FullName,
		student.PhoneNumber,
		student.ParentName,
		student.ParentPhoneNumber,
		student.CuratorId,
		student.PlatformLink,
		student.CrmLink,
		student.CreatedAt,
		student.IsActive,
	)

	err := row.Scan(&student.Id)
	if err != nil {
		return uuid.UUID{}, err
	}

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
        s.platform_link, 
        s.crm_link, 
        s.created_at,
        s.is_active
    FROM students s
    WHERE 1=1`
    
    params := pgx.NamedArgs{}

    if filters.Search != "" {
        sql += " AND s.full_name ILIKE @search"
        params["search"] = "%" + filters.Search + "%"
    }

    if filters.Course != "" {
        courseUUID, err := uuid.Parse(filters.Course)
        if err != nil {
            return nil, err
        }
        sql += " AND s.course_id = @course"
        params["course"] = courseUUID
    }

    if filters.IsActive != "" {
        sql += " AND s.is_active = @is_active"
        params["is_active"] = filters.IsActive
    }

    if filters.CuratorId != "" {
        curatorUUID, err := uuid.Parse(filters.CuratorId)
        if err != nil {
            return nil, err
        }
        sql += " AND s.curator_id = @curator_id"
        params["curator_id"] = curatorUUID
    }

    rows, err := r.db.Query(c, sql, params)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

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
			s.platform_link, 
			s.crm_link, 
			s.created_at,
			s.is_active 
			FROM students s
			WHERE s.id = $1`

	var student models.Student
	row := r.db.QueryRow(c, sql, studentId)
	err := row.Scan(
		&studentId,
		&student.CourseId,
		&student.FullName,
		&student.PhoneNumber,
		&student.ParentName,
		&student.ParentPhoneNumber,
		&student.CuratorId,
		&student.PlatformLink,
		&student.CrmLink,
		&student.CreatedAt,
		&student.IsActive,
	)
	if err != nil {
		return models.Student{}, err
	}
	return student, nil
}

func (r *StudentsRepository) Update(c context.Context, student models.Student) error {
    tx, err := r.db.Begin(c)
    if err != nil {
        return err
    }
    defer func() {
        if err != nil {
            tx.Rollback(c)
        }
    }()

    _, err = tx.Exec(c, `
    UPDATE students SET
        course_id = $1,
        full_name = $2,
        phone_number = $3,
        parent_name = $4,
        parent_phone_number = $5,
        curator_id = $6,
        platform_link = $7,
        crm_link = $8,
        created_at = $9,
        is_active = $10
    WHERE id = $11`,
        student.CourseId,
        student.FullName,
        student.PhoneNumber,
        student.ParentName,
        student.ParentPhoneNumber,
        student.CuratorId,
        student.PlatformLink,
        student.CrmLink,
        student.CreatedAt,
        student.IsActive,
        student.Id)

    if err != nil {
        return err
    }

    return tx.Commit(c)
}

func (r *StudentsRepository) Delete(c context.Context, studentId uuid.UUID) error {
	tx, err := r.db.Begin(c)
	if err != nil {
		return err
	}

	_, err = tx.Exec(c, "DELETE FROM students WHERE id = $1", studentId)
	if err != nil {
		tx.Rollback(c)
		return err
	}

	err = tx.Commit(c)
	if err != nil {
		return err
	}
	return nil
}
