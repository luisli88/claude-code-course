package persistence

import (
	"database/sql"
	"errors"
	"myapp/internal/domain/entity"
	"myapp/internal/domain/repository"
	"time"
)

var _ repository.PasswordResetTokenRepository = (*PostgresPasswordResetTokenRepo)(nil)

// PostgresPasswordResetTokenRepo implements PasswordResetTokenRepository against a PostgreSQL database.
type PostgresPasswordResetTokenRepo struct {
	db *sql.DB
}

// NewPostgresPasswordResetTokenRepo creates a new PostgresPasswordResetTokenRepo.
func NewPostgresPasswordResetTokenRepo(db *sql.DB) *PostgresPasswordResetTokenRepo {
	return &PostgresPasswordResetTokenRepo{db: db}
}

// Create persists a new token and returns the record with DB-generated id and created_at.
func (r *PostgresPasswordResetTokenRepo) Create(token entity.PasswordResetToken) (*entity.PasswordResetToken, error) {
	var created entity.PasswordResetToken
	err := r.db.QueryRow(
		`INSERT INTO password_reset_tokens (user_id, token, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, token, expires_at, used, created_at`,
		token.UserID, token.Token, token.ExpiresAt,
	).Scan(&created.ID, &created.UserID, &created.Token, &created.ExpiresAt, &created.Used, &created.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

// FindByToken retrieves a token by its string value. Returns ErrTokenNotFound if absent.
func (r *PostgresPasswordResetTokenRepo) FindByToken(token string) (*entity.PasswordResetToken, error) {
	var t entity.PasswordResetToken
	err := r.db.QueryRow(
		`SELECT id, user_id, token, expires_at, used, created_at
		 FROM password_reset_tokens
		 WHERE token = $1`,
		token,
	).Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.Used, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrTokenNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// MarkAsUsed sets used = TRUE for the token with the given id.
func (r *PostgresPasswordResetTokenRepo) MarkAsUsed(id int) error {
	_, err := r.db.Exec(
		`UPDATE password_reset_tokens SET used = TRUE WHERE id = $1`,
		id,
	)
	return err
}

// CountRecentByEmail counts tokens for the user with the given email created at or after since.
func (r *PostgresPasswordResetTokenRepo) CountRecentByEmail(email string, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		 FROM password_reset_tokens prt
		 JOIN users u ON u.id = prt.user_id
		 WHERE u.email = $1 AND prt.created_at >= $2`,
		email, since,
	).Scan(&count)
	return count, err
}

// InvalidatePreviousByUserID marks all unused tokens for the given user as used.
func (r *PostgresPasswordResetTokenRepo) InvalidatePreviousByUserID(userID string) error {
	_, err := r.db.Exec(
		`UPDATE password_reset_tokens SET used = TRUE WHERE user_id = $1 AND used = FALSE`,
		userID,
	)
	return err
}
