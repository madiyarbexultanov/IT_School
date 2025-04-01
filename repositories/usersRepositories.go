package repositories

import (
	"context"
	"it_school/models"
	"time"

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

func (r *UsersRepository) FindByEmail(c context.Context, email string) (models.User, error) {
	var user models.User
	row := r.db.QueryRow(c, "select id, email, password from users where email = $1", email)
	if err := row.Scan(&user.Id, &user.Email, &user.PasswordHash); err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (r *UsersRepository) SetResetToken(c context.Context, email, resetToken string, expirationTime time.Time) error {
	query := `UPDATE users SET reset_token = $1, reset_token_expires_at = $2 WHERE email = $3`
	_, err := r.db.Exec(c, query, resetToken, expirationTime, email)
	return err
}

func (r *UsersRepository) GetUserByResetToken(c context.Context, resetToken string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email FROM users WHERE reset_token = $1 AND reset_token_expires_at > NOW()`
	err := r.db.QueryRow(c, query, resetToken).Scan(&user.Id, &user.Email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UsersRepository) ClearResetToken(c context.Context, userID int) error {
	query := `UPDATE users SET reset_token = NULL, reset_token_expires_at = NULL WHERE id = $1`
	_, err := r.db.Exec(c, query, userID)
	return err
}

func (r *UsersRepository) UpdatePassword(c context.Context, userID int, hashedPassword string) error {
	query := `UPDATE users SET password = $1, reset_token = NULL WHERE id = $2`
	_, err := r.db.Exec(c, query, hashedPassword, userID)
	return err
}
