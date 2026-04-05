package usecase_test

import (
	"errors"
	"testing"
	"time"

	"myapp/internal/application/usecase"
	"myapp/internal/domain/entity"
	"myapp/internal/domain/repository"
)

func TestRegister_Execute(t *testing.T) {
	validInput := usecase.RegisterInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "secret123",
	}

	t.Run("creates user successfully", func(t *testing.T) {
		mock := &repository.MockUserRepository{}
		uc := usecase.NewRegister(mock)

		output, err := uc.Execute(validInput)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output.User.Name != validInput.Name {
			t.Errorf("expected name %q, got %q", validInput.Name, output.User.Name)
		}
		if output.User.Email != validInput.Email {
			t.Errorf("expected email %q, got %q", validInput.Email, output.User.Email)
		}
		if output.User.ID == "" {
			t.Error("expected non-empty ID")
		}
		if output.User.PasswordHash == "" {
			t.Error("expected non-empty password hash")
		}
		if output.User.PasswordHash == validInput.Password {
			t.Error("password must be hashed, not stored in plain text")
		}
		if output.User.CreatedAt.IsZero() {
			t.Error("expected non-zero created_at")
		}
	})

	t.Run("returns error when email already taken", func(t *testing.T) {
		existing := entity.User{
			ID:        "1",
			Name:      "Alice",
			Email:     "alice@example.com",
			CreatedAt: time.Now(),
		}
		mock := &repository.MockUserRepository{Users: []entity.User{existing}}
		uc := usecase.NewRegister(mock)

		_, err := uc.Execute(validInput)
		if !errors.Is(err, usecase.ErrEmailAlreadyTaken) {
			t.Errorf("expected ErrEmailAlreadyTaken, got %v", err)
		}
	})

	t.Run("returns invalid input when name is empty", func(t *testing.T) {
		mock := &repository.MockUserRepository{}
		uc := usecase.NewRegister(mock)

		_, err := uc.Execute(usecase.RegisterInput{Name: "", Email: "a@b.com", Password: "secret123"})
		if !errors.Is(err, usecase.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("returns invalid input when email has no @", func(t *testing.T) {
		mock := &repository.MockUserRepository{}
		uc := usecase.NewRegister(mock)

		_, err := uc.Execute(usecase.RegisterInput{Name: "Alice", Email: "notanemail", Password: "secret123"})
		if !errors.Is(err, usecase.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("returns invalid input when password is too short", func(t *testing.T) {
		mock := &repository.MockUserRepository{}
		uc := usecase.NewRegister(mock)

		_, err := uc.Execute(usecase.RegisterInput{Name: "Alice", Email: "a@b.com", Password: "short"})
		if !errors.Is(err, usecase.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("propagates repository error on Create", func(t *testing.T) {
		dbErr := errors.New("db down")
		uc := usecase.NewRegister(&createErrorMock{err: dbErr})

		_, err := uc.Execute(validInput)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// createErrorMock returns ErrNotFound on FindByEmail but fails on Create.
type createErrorMock struct {
	err error
}

func (m *createErrorMock) FindAll(_ repository.PaginationParams) ([]entity.User, error) {
	return nil, nil
}
func (m *createErrorMock) Count() (int, error) { return 0, nil }
func (m *createErrorMock) FindByEmail(_ string) (*entity.User, error) {
	return nil, repository.ErrNotFound
}
func (m *createErrorMock) Create(_ entity.User) (*entity.User, error) {
	return nil, m.err
}
