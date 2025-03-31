package repositories

import (
	"context"
	"it_school/models"


	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionsRepository struct {
	db *pgxpool.Pool
}

func NewSessionsRepository(conn *pgxpool.Pool) *SessionsRepository {
	return &SessionsRepository{db: conn}
}

func (r *SessionsRepository) CreateSession(ctx context.Context, session models.Session) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO sessions (user_id, refresh_token, expires_at) 
		 VALUES ($1, $2, $3)`,
		session.UserID, session.RefreshToken, session.ExpiresAt)
	return err
}

func (r *SessionsRepository) GetSession(ctx context.Context, refreshToken string) (models.Session, error) {
	var session models.Session
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, refresh_token, expires_at 
		 FROM sessions 
		 WHERE refresh_token = $1 AND expires_at > NOW()`,
		refreshToken).
		Scan(&session.ID, &session.UserID, &session.RefreshToken, &session.ExpiresAt)
	return session, err
}

func (r *SessionsRepository) DeleteSession(ctx context.Context, refreshToken string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM sessions 
		 WHERE refresh_token = $1`,
		refreshToken)
	return err
}