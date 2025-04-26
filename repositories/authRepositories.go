package repositories

import (
	"context"
	"it_school/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(conn *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{db: conn}
}

func (r *AuthRepository) SetResetToken(c context.Context, email, resetToken string, expirationTime time.Time) error {
	query := `UPDATE users SET reset_token = $1, reset_token_expires_at = $2 WHERE email = $3`
	_, err := r.db.Exec(c, query, resetToken, expirationTime, email)
	return err
}

func (r *AuthRepository) GetUserByResetToken(c context.Context, resetToken string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email FROM users WHERE reset_token = $1 AND reset_token_expires_at > NOW()`
	err := r.db.QueryRow(c, query, resetToken).Scan(&user.Id, &user.Email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) ClearResetToken(c context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET reset_token = NULL, reset_token_expires_at = NULL WHERE id = $1`
	_, err := r.db.Exec(c, query, userID)
	return err
}

func (r *AuthRepository) UpdatePassword(c context.Context, userID uuid.UUID, hashedPassword string) error {
	query := `UPDATE users SET password = $1, reset_token = NULL WHERE id = $2`
	_, err := r.db.Exec(c, query, hashedPassword, userID)
	return err
}


