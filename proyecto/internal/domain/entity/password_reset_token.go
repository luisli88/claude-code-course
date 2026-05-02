package entity

import "time"

// PasswordResetToken represents a one-time-use token for password reset via email.
type PasswordResetToken struct {
	ID        int
	UserID    string
	Token     string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

// IsExpired reports whether the token has passed its expiration time.
func (t PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid reports whether the token can still be used to reset a password.
func (t PasswordResetToken) IsValid() bool {
	return !t.Used && !t.IsExpired()
}
