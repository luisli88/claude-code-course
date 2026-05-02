package repository

import (
	"myapp/internal/domain/entity"
	"time"
)

var _ PasswordResetTokenRepository = (*MockPasswordResetTokenRepository)(nil)

// MockPasswordResetTokenRepository is an in-memory implementation of PasswordResetTokenRepository for unit tests.
type MockPasswordResetTokenRepository struct {
	// Tokens holds all stored tokens and is directly readable in tests for assertions.
	Tokens []entity.PasswordResetToken
	// Err, when non-nil, is returned by every method to simulate failures.
	Err error
	// EmailToUserID maps an email address to a user ID so CountRecentByEmail can resolve tokens.
	EmailToUserID map[string]string
	nextID        int
}

// Create stores the token with a generated ID and current timestamp.
func (m *MockPasswordResetTokenRepository) Create(token entity.PasswordResetToken) (*entity.PasswordResetToken, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	m.nextID++
	token.ID = m.nextID
	token.CreatedAt = time.Now()
	m.Tokens = append(m.Tokens, token)
	return &token, nil
}

// FindByToken searches the in-memory slice by token string. Returns ErrTokenNotFound if absent.
func (m *MockPasswordResetTokenRepository) FindByToken(token string) (*entity.PasswordResetToken, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	for i := range m.Tokens {
		if m.Tokens[i].Token == token {
			return &m.Tokens[i], nil
		}
	}
	return nil, ErrTokenNotFound
}

// MarkAsUsed sets Used = true on the token with the given ID. Returns ErrTokenNotFound if absent.
func (m *MockPasswordResetTokenRepository) MarkAsUsed(id int) error {
	if m.Err != nil {
		return m.Err
	}
	for i := range m.Tokens {
		if m.Tokens[i].ID == id {
			m.Tokens[i].Used = true
			return nil
		}
	}
	return ErrTokenNotFound
}

// CountRecentByEmail counts tokens for the user resolved via EmailToUserID that were created at or after since.
func (m *MockPasswordResetTokenRepository) CountRecentByEmail(email string, since time.Time) (int, error) {
	if m.Err != nil {
		return 0, m.Err
	}
	userID, ok := m.EmailToUserID[email]
	if !ok {
		return 0, nil
	}
	count := 0
	for _, t := range m.Tokens {
		if t.UserID == userID && !t.CreatedAt.Before(since) {
			count++
		}
	}
	return count, nil
}

// InvalidatePreviousByUserID marks all unused tokens for the given user as used.
func (m *MockPasswordResetTokenRepository) InvalidatePreviousByUserID(userID string) error {
	if m.Err != nil {
		return m.Err
	}
	for i := range m.Tokens {
		if m.Tokens[i].UserID == userID && !m.Tokens[i].Used {
			m.Tokens[i].Used = true
		}
	}
	return nil
}
