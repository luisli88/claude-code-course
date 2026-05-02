package repository

import (
	"errors"
	"myapp/internal/domain/entity"
	"time"
)

var (
	ErrTokenNotFound    = errors.New("password reset token not found")
	ErrTokenExpired     = errors.New("password reset token has expired")
	ErrTokenAlreadyUsed = errors.New("password reset token has already been used")
)

// PasswordResetTokenRepository defines the persistence contract for password reset tokens.
type PasswordResetTokenRepository interface {
	// Create persists a new password reset token and returns the stored record with DB-generated fields.
	Create(token entity.PasswordResetToken) (*entity.PasswordResetToken, error)

	// FindByToken retrieves a token by its string value. Returns ErrTokenNotFound if it doesn't exist.
	FindByToken(token string) (*entity.PasswordResetToken, error)

	// MarkAsUsed marks the token identified by id as used, preventing reuse.
	MarkAsUsed(id int) error

	// CountRecentByEmail returns the number of tokens created for the user with the given email since since.
	CountRecentByEmail(email string, since time.Time) (int, error)

	// InvalidatePreviousByUserID marks all unused tokens for the given user as used.
	InvalidatePreviousByUserID(userID string) error
}
