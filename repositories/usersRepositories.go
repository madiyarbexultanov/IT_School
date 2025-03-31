package repositories

import (
	"context"
	"it_school/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersRepository struct {
	db *pgxpool.Pool
}

func NewUsersRepository(conn *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{db: conn}
}

func (r *UsersRepository) FindAll(c context.Context) ([]models.User, error)  {
	sql := "select id, email from users where 1=1"

	rows, err := r.db.Query(c, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.Id, &user.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UsersRepository) FindById(c context.Context, id int) (models.User, error)  {
	var user models.User
	row := r.db.QueryRow(c, "select id, email, role_id from users where id = $1", id)
	err := row.Scan(&user.Id, &user.Email, &user.RoleID)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *UsersRepository) ChangePasswordHash(c context.Context, id int, password string) error {
	_, err := r.db.Exec(c, "update users set password=$1 where id=$2", password, id)
	if err != nil {
		return err
	}
	return nil
}


func (r *UsersRepository) FindByEmail(c context.Context, email string) (models.User, error) {
	var user models.User
	row := r.db.QueryRow(c, "select id, email, password from users where email = $1", email)
	if err := row.Scan(&user.Id, &user.Email, &user.PasswordHash); err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (r *UsersRepository) AssignRole(c context.Context, userID int, roleID int) error {
	_, err := r.db.Exec(c, "UPDATE users SET role_id = $1 WHERE id = $2", roleID, userID)
	return err
}