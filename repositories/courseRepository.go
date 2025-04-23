package repositories

import (
	"context"
	"it_school/logger"
	"it_school/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type CourseRepository struct {
	db *pgxpool.Pool
}

func NewCourseRepository(conn *pgxpool.Pool) *CourseRepository {
	return &CourseRepository{db: conn}
}



func (r *CourseRepository) Create(c context.Context, course models.Course) (uuid.UUID, error) {
	course.Id = uuid.New()
	row := r.db.QueryRow(c, `insert into courses (id, title) values ($1, $2) returning id`, course.Id, course.Title)
	err := row.Scan(&course.Id)
	if err != nil {
		return uuid.UUID{}, err
	}
	return course.Id, nil
}

func (r *CourseRepository) Update(c context.Context, updateCourse models.Course) error {
	l := logger.GetLogger()

	tx, err := r.db.Begin(c)
	if err != nil {
		l.Error("Ошибка начала транзакции", zap.String("db_msg", err.Error()))
		return err
	}
	_, err = tx.Exec(c, `update courses set title = $1 where id = $2`, updateCourse.Title, updateCourse.Id)
	if err != nil {
		return err
	}

	err = tx.Commit(c)
	if err != nil {
		l.Error("Ошибка при коммите транзакции", zap.String("commit_msg", err.Error()))
		return err
	}
	return nil
}

func (r *CourseRepository) FindAll(c context.Context) ([]models.Course, error) {
	sql := `select c.id, c.title from courses c`

	row, err := r.db.Query(c, sql)
	if err != nil {
		return nil, err
	}

	courses := make([]models.Course, 0)
	for row.Next() {
		var course models.Course
		err := row.Scan(&course.Id, &course.Title)
		if err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}
	return courses, nil
}

func (r *CourseRepository) FindById(c context.Context, courseId uuid.UUID) (models.Course, error) {
	var course models.Course
	row := r.db.QueryRow(c, `select c.id, c.title from courses c where c.id = $1`, courseId)
	if err := row.Scan(&course.Id, &course.Title); err != nil {
		return models.Course{}, err
	}
	return course, nil
}

func (r *CourseRepository) Delete(c context.Context, courseId uuid.UUID) error {
	l := logger.GetLogger()

	tx, err := r.db.Begin(c)
	if err != nil {
		l.Error("Ошибка начала транзакции", zap.String("db_msg", err.Error()))
		return err
	}

	_, err = tx.Exec(c, "DELETE FROM courses WHERE id = $1", courseId)
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
