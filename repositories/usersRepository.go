package repositories

import (
	"context"
	"it_school/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersRepository struct {
	db *pgxpool.Pool
}

func NewRUsersRepository(conn *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{db: conn}
}


func (r *UsersRepository) FindAll(c context.Context, roleID *uuid.UUID) ([]models.User, error) {
	var (
		rows pgx.Rows
		err  error
	)

	if roleID != nil {
		query := `
			SELECT id, full_name, email, phone_number, role_id
			FROM users
			WHERE role_id = $1;
		`
		rows, err = r.db.Query(c, query, *roleID)
	} else {
		query := `
			SELECT id, full_name, email, phone_number, role_id
			FROM users;
		`
		rows, err = r.db.Query(c, query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.Id, &u.Full_name, &u.Email, &u.Telephone, &u.RoleID); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}




func (r *UsersRepository) FindById(c context.Context, id uuid.UUID) (models.User, error) {
	var user models.User
	row := r.db.QueryRow(c, "select id, email, full_name, phone_number, role_id from users where id=$1", id)
	err := row.Scan(&user.Id, &user.Email, &user.Full_name, &user.Telephone, &user.RoleID)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *UsersRepository) FindByEmail(c context.Context, email string) (models.User, error) {
	var user models.User
	row := r.db.QueryRow(c, "select id, email, password, role_id from users where email = $1", email)
	if err := row.Scan(&user.Id, &user.Email, &user.PasswordHash, &user.RoleID); err != nil {
		return models.User{}, err
	}

	return user, nil
}


func (r *UsersRepository) Create(c context.Context, user models.User) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(c, "insert into users(email, password, full_name, phone_number, role_id) values($1, $2, $3, $4, $5) returning id",
							user.Email, user.PasswordHash, user.Full_name, user.Telephone, user.RoleID).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *UsersRepository) Update(c context.Context, id uuid.UUID, user models.User) error {
	_, err := r.db.Exec(c, `
	UPDATE users SET email=$1, full_name=$2, phone_number=$3 WHERE id=$4`,
	 user.Email, user.Full_name, user.Telephone, id)

	if err != nil {
		return err
	}
	return nil
}

func (r *UsersRepository) Delete(c context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(c, "delete from users where id=$1", id)
	if err != nil {
		return err
	}
	return nil
}

func (r *UsersRepository) CountByRoleID(ctx context.Context, roleID uuid.UUID) (int, error) {
    var cnt int
    err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role_id = $1`, roleID).Scan(&cnt)
    return cnt, err
}