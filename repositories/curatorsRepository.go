package repositories

import (
	"context"
	"it_school/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CuratorsRepository struct {
	db *pgxpool.Pool
}

func NewCuratorsRepository(conn *pgxpool.Pool) *CuratorsRepository {
	return &CuratorsRepository{db: conn}
}

type CuratorResponse struct {
    ID         uuid.UUID   `json:"id"`
    FullName   string      `json:"full_name"`
    Email      string      `json:"email"`
    Telephone  string      `json:"telephone"`
    RoleID     uuid.UUID   `json:"role_id"`
    StudentIDs []uuid.UUID `json:"student_ids"`
    CourseIDs  []uuid.UUID `json:"course_ids"`
}

func (r *CuratorsRepository) GetCuratorByUserID(ctx context.Context, userID uuid.UUID) (models.Curator, error) {
	var curator models.Curator
	query := `SELECT user_id, student_ids, course_ids FROM curators WHERE user_id = $1`
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&curator.UserID,
		&curator.StudentIDs,
		&curator.CourseIDs,
	)
	return curator, err
}

func (r *CuratorsRepository) Create(c context.Context, curator models.Curator) error {
	_, err := r.db.Exec(c,
		`INSERT INTO curators(user_id, student_ids, course_ids) VALUES ($1, $2, $3)`,
		curator.UserID, curator.StudentIDs, curator.CourseIDs,
	)
	return err
}

func (r *CuratorsRepository) AddStudent(c context.Context, curatorID, studentID uuid.UUID) error {
	tx, err := r.db.Begin(c)
	if err != nil {
		return err
	}
	defer tx.Rollback(c)

	// 1. Добавляем студента в curators
	_, err = tx.Exec(c,
		`UPDATE curators SET student_ids = array_append(student_ids, $1) 
		 WHERE user_id = $2 AND NOT ($1 = ANY(student_ids))`,
		studentID, curatorID,
	)
	if err != nil {
		return err
	}

	// 2. Обновляем curator_id у студента
	_, err = tx.Exec(c,
		`UPDATE students SET curator_id = $1 WHERE id = $2`,
		curatorID, studentID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(c)
}


func (r *CuratorsRepository) AddCourse(c context.Context, curatorID, courseID uuid.UUID) error {
	_, err := r.db.Exec(c,
		`UPDATE curators SET course_ids = array_append(course_ids, $1) WHERE user_id = $2 AND NOT ($1 = ANY(course_ids))`,
		courseID, curatorID,
	)
	return err
}

func (r *CuratorsRepository) RemoveStudent(c context.Context, curatorID, studentID uuid.UUID) error {
	tx, err := r.db.Begin(c)
	if err != nil {
		return err
	}
	defer tx.Rollback(c)

	// 1. Убираем студента из массива у куратора
	_, err = tx.Exec(c,
		`UPDATE curators SET student_ids = array_remove(student_ids, $1) 
		 WHERE user_id = $2`,
		studentID, curatorID,
	)
	if err != nil {
		return err
	}

	// 2. Обнуляем curator_id у студента
	_, err = tx.Exec(c,
		`UPDATE students SET curator_id = NULL WHERE id = $1 AND curator_id = $2`,
		studentID, curatorID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(c)
}

func (r *CuratorsRepository) RemoveCourse(c context.Context, curatorID, courseID uuid.UUID) error {
	_, err := r.db.Exec(c,
		`UPDATE curators SET course_ids = array_remove(course_ids, $1) WHERE user_id = $2`,
		courseID, curatorID,
	)
	return err
}
